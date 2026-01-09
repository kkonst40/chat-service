package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/kkonst40/ichat/internal/apperror"
	"github.com/kkonst40/ichat/internal/logger"
	"github.com/kkonst40/ichat/internal/model"
	"github.com/kkonst40/ichat/internal/repository"
)

type MessageService struct {
	messageRepository repository.MessageRepository
	chatService       *ChatService
	userService       *UserService
}

func NewMessageService(
	newMessageRepository repository.MessageRepository,
	newChatService *ChatService,
	newUserService *UserService,
) *MessageService {
	service := MessageService{
		messageRepository: newMessageRepository,
		chatService:       newChatService,
		userService:       newUserService,
	}

	return &service
}

func (s *MessageService) GetChatMessages(ctx context.Context, chatID uuid.UUID, requesterID uuid.UUID) ([]model.Message, error) {
	log := logger.FromContext(ctx)
	log.Debug("messageService.GetChatMessages", "chatID", chatID)

	if !s.chatService.DoesChatExist(ctx, chatID) {
		return nil, &apperror.NotFoundError{Msg: fmt.Sprintf("chat (%v) not found", chatID)}
	}
	if !s.userService.isUserInChat(ctx, chatID, requesterID) {
		return nil, &apperror.ForbiddenError{Msg: fmt.Sprintf("user (%v) is not in the chat (%v)", requesterID, chatID)}
	}

	messages, err := s.messageRepository.GetChatMessages(ctx, chatID)
	if err != nil {
		return nil, err
	}
	log.Debug("messages retrieved")

	return messages, nil
}

func (s *MessageService) CreateMessage(ctx context.Context, msgID, userID, chatID uuid.UUID, text string) (*model.Message, error) {
	log := logger.FromContext(ctx)
	log.Debug("messageService.CreateMessage", "chatID", chatID)

	if !s.chatService.DoesChatExist(ctx, chatID) {
		return nil, &apperror.NotFoundError{Msg: fmt.Sprintf("chat (%v) not found", chatID)}
	}

	message := &model.Message{
		ID:        msgID,
		UserID:    userID,
		ChatID:    chatID,
		Text:      text,
		CreatedAt: time.Now(),
	}

	if err := s.messageRepository.CreateMessage(ctx, message); err != nil {
		return nil, err
	}
	log.Debug("message created")

	return message, nil
}

func (s *MessageService) UpdateMessage(ctx context.Context, msgID uuid.UUID, text string, requesterID uuid.UUID) error {
	log := logger.FromContext(ctx)
	log.Debug("messageService.UpdateMessage", "msgID", msgID)

	message, err := s.messageRepository.GetMessage(ctx, msgID)
	if err != nil {
		return err
	}

	if message.UserID != requesterID {
		return &apperror.ForbiddenError{Msg: fmt.Sprintf("user (%v) is not the owner of the message (%v)", requesterID, msgID)}
	}

	newMessage := &model.Message{
		ID:        msgID,
		UserID:    message.UserID,
		ChatID:    message.ChatID,
		Text:      text,
		CreatedAt: message.CreatedAt,
	}

	err = s.messageRepository.UpdateMessage(ctx, newMessage)
	if err != nil {
		return err
	}
	log.Debug("message updated")

	return nil
}

func (s *MessageService) DeleteMessage(ctx context.Context, msgID uuid.UUID, requesterID uuid.UUID) error {
	log := logger.FromContext(ctx)
	log.Debug("messageService.DeleteMessage", "msgID", msgID)

	message, err := s.messageRepository.GetMessage(ctx, msgID)
	if err != nil {
		return err
	}

	if message.UserID == requesterID {
		err = s.messageRepository.DeleteMessage(ctx, msgID)
		if err != nil {
			return err
		}
		log.Debug("message deleted")

		return nil
	}

	sender, err := s.userService.GetChatUser(ctx, message.ChatID, message.UserID)
	if err != nil {
		return err
	}
	requester, err := s.userService.GetChatUser(ctx, message.ChatID, requesterID)
	if err != nil {
		return err
	}

	if !(model.UserPriority[requester.Role] > model.UserPriority[sender.Role]) {
		return &apperror.ForbiddenError{Msg: fmt.Sprintf("user (%v) has no permission", requesterID)}
	}

	err = s.messageRepository.DeleteMessage(ctx, msgID)
	if err != nil {
		return err
	}
	log.Debug("message deleted")

	return nil
}
