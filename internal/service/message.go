package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/kkonst40/ichat/internal/dispatcher"
	errs "github.com/kkonst40/ichat/internal/domain/errors"
	"github.com/kkonst40/ichat/internal/domain/event"
	"github.com/kkonst40/ichat/internal/domain/model"
	"github.com/kkonst40/ichat/internal/repository"
)

type MessageService struct {
	messageRepository repository.MessageRepository
	chatService       *ChatService
	userService       *UserService
	dispatcher        *dispatcher.Dispatcher
	textMaxLength     int
}

func NewMessageService(
	messageRepository repository.MessageRepository,
	chatService *ChatService,
	userService *UserService,
	dispatcher *dispatcher.Dispatcher,
	textMaxLength int,
) *MessageService {
	service := MessageService{
		messageRepository: messageRepository,
		chatService:       chatService,
		userService:       userService,
		dispatcher:        dispatcher,
		textMaxLength:     textMaxLength,
	}

	return &service
}

func (s *MessageService) GetChatMessages(ctx context.Context, chatID uuid.UUID, from uuid.UUID, count int64, requesterID uuid.UUID) ([]model.Message, error) {
	slog.DebugContext(ctx, "messageService.GetChatMessages", "chatID", chatID)

	if !s.chatService.ChatExists(ctx, chatID) {
		return nil, fmt.Errorf(
			"%w: chat %v",
			errs.ErrChatNotFound,
			chatID,
		)
	}
	if !s.userService.userInChat(ctx, chatID, requesterID) {
		return nil, fmt.Errorf(
			"%w: user %v is not in chat %v",
			errs.ErrForbidden,
			requesterID,
			chatID,
		)
	}

	messages, err := s.messageRepository.GetChatMessages(ctx, chatID, from, count)
	if err != nil {
		return nil, fmt.Errorf("get chat %v messages: %w", chatID, err)
	}
	slog.DebugContext(ctx, "messages retrieved")

	userIDs := make([]uuid.UUID, 0, len(messages))
	for i := range messages {
		userIDs = append(userIDs, messages[i].UserID)
	}

	logins, err := s.userService.getUserLogins(ctx, userIDs)
	if err != nil {
		return nil, fmt.Errorf("get user logins: %w", err)
	}

	for i := range messages {
		messages[i].UserName = logins[messages[i].UserID]
	}

	return messages, nil
}

func (s *MessageService) CreateMessage(ctx context.Context, userID, chatID uuid.UUID, text string) (*model.Message, error) {
	slog.DebugContext(ctx, "messageService.CreateMessage", "chatID", chatID)

	newID, err := uuid.NewV7()
	if err != nil {
		return nil, fmt.Errorf("%w: generating uuid: %w", errs.ErrInternal, err)
	}

	msg := &model.Message{
		ID:        newID,
		UserID:    userID,
		ChatID:    chatID,
		Text:      limitText(text, s.textMaxLength),
		CreatedAt: time.Now(),
	}

	if err := s.messageRepository.CreateMessage(ctx, msg); err != nil {
		if errors.Is(err, errs.ErrChatNotFound) {
			return nil, fmt.Errorf("%w: ID %v", err, chatID)
		}

		return nil, fmt.Errorf("create message: %w", err)
	}
	slog.DebugContext(ctx, "message created")

	nameMap, err := s.userService.getUserLogins(ctx, []uuid.UUID{userID})
	if err != nil {
		return nil, fmt.Errorf("get user logins: %w", err)
	}
	userName := nameMap[userID]

	s.dispatcher.Publish(event.Event{
		Type:   event.CreateMsg,
		ChatID: chatID,
		Payload: event.CreateMsgEvent{
			MsgID:    newID,
			UserID:   userID,
			UserName: userName,
			Text:     msg.Text,
		},
	})

	return msg, nil
}

func (s *MessageService) UpdateMessage(ctx context.Context, msgID uuid.UUID, text string, requesterID uuid.UUID) error {
	slog.DebugContext(ctx, "messageService.UpdateMessage", "msgID", msgID)

	msg, err := s.messageRepository.GetMessage(ctx, msgID)
	if err != nil {
		if errors.Is(err, errs.ErrMsgNotFound) {
			return fmt.Errorf("%w: ID %v", err, msgID)
		}
		return fmt.Errorf("get message %v to update: %w", msgID, err)
	}

	if msg.UserID != requesterID {
		return fmt.Errorf(
			"%w: user %v has no permission to update message %v",
			errs.ErrForbidden,
			requesterID,
			msgID,
		)
	}

	newMsg := &model.Message{
		ID:        msgID,
		UserID:    msg.UserID,
		ChatID:    msg.ChatID,
		Text:      limitText(text, s.textMaxLength),
		CreatedAt: msg.CreatedAt,
	}

	err = s.messageRepository.UpdateMessage(ctx, newMsg)
	if err != nil {
		if errors.Is(err, errs.ErrMsgNotFound) {
			return fmt.Errorf("%w: ID %v", err, msgID)
		}
		return fmt.Errorf("update message %v: %w", msgID, err)
	}
	slog.DebugContext(ctx, "message updated")

	s.dispatcher.Publish(event.Event{
		Type:   event.UpdateMsg,
		ChatID: msg.ChatID,
		Payload: event.UpdateMsgEvent{
			MsgID: msgID,
			Text:  newMsg.Text,
		},
	})

	return nil
}

func (s *MessageService) DeleteMessage(ctx context.Context, msgID uuid.UUID, requesterID uuid.UUID) error {
	slog.DebugContext(ctx, "messageService.DeleteMessage", "msgID", msgID)

	msg, err := s.messageRepository.GetMessage(ctx, msgID)
	if err != nil {
		if errors.Is(err, errs.ErrMsgNotFound) {
			return fmt.Errorf("%w: ID %v", err, msgID)
		}
		return fmt.Errorf("get message %v to delete: %w", msgID, err)
	}

	if msg.UserID == requesterID {
		err = s.messageRepository.DeleteMessage(ctx, msgID)
		if err != nil {
			return fmt.Errorf("delete message %v: %w", msgID, err)
		}
		slog.DebugContext(ctx, "message deleted")

		return nil
	}

	sender, err := s.userService.GetChatUser(ctx, msg.ChatID, msg.UserID)
	if err != nil {
		return fmt.Errorf("get user %v to delete his message: %w", msg.UserID, err)
	}
	requester, err := s.userService.GetChatUser(ctx, msg.ChatID, requesterID)
	if err != nil {
		return fmt.Errorf("get requester %v to delete message: %w", requesterID, err)
	}

	if !(model.UserPriority[requester.Role] > model.UserPriority[sender.Role]) {
		return fmt.Errorf(
			"%w: user %v has no permission to delete message %v",
			errs.ErrForbidden,
			requesterID,
			msgID,
		)
	}

	err = s.messageRepository.DeleteMessage(ctx, msgID)
	if err != nil {
		return fmt.Errorf("delete message %v: %w", msgID, err)
	}
	slog.DebugContext(ctx, "message deleted")

	s.dispatcher.Publish(event.Event{
		Type:   event.DeleteMsg,
		ChatID: msg.ChatID,
		Payload: event.DeleteMsgEvent{
			MsgID: msgID,
		},
	})

	return nil
}

func limitText(s string, maxChars int) string {
	if len(s) <= maxChars {
		return s
	}

	count := 0
	for byteIndex := range s {
		if count == maxChars {
			return s[:byteIndex]
		}
		count++
	}
	return s
}
