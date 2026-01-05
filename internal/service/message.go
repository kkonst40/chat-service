package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
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

func (s *MessageService) GetChatMessages(ctx context.Context, chatID uuid.UUID, requesterID uuid.UUID) ([]*model.Message, error) {
	log := logger.FromContext(ctx)
	log.Debug("messageService.GetChatMessages", "chatID", chatID)

	if !s.chatService.doesChatExist(ctx, chatID) {
		return nil, fmt.Errorf("chat does not exist")
	}
	if !s.userService.isUserInChat(ctx, chatID, requesterID) {
		return nil, fmt.Errorf("user is not in the chat")
	}

	chats, err := s.messageRepository.GetChatMessages(ctx, chatID)
	if err != nil {
		return nil, err
	}
	log.Debug("messages retrieved")

	return chats, nil
}

func (s *MessageService) CreateMessage(ctx context.Context, userID, chatID uuid.UUID, text string) (*model.Message, error) {
	log := logger.FromContext(ctx)
	log.Debug("messageService.CreateMessage", "chatID", chatID)

	if !s.chatService.doesChatExist(ctx, chatID) {
		return nil, fmt.Errorf("chat does not exist")
	}

	newID, err := uuid.NewV7()
	if err != nil {
		return nil, err
	}

	message := &model.Message{
		ID:        newID,
		UserID:    userID,
		ChatID:    chatID,
		Text:      text,
		CreatedAt: time.Now(),
	}

	if err = s.messageRepository.CreateMessage(ctx, message); err != nil {
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
		return fmt.Errorf("user is not the owner of the message")
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
		return fmt.Errorf("user has no permission")
	}

	err = s.messageRepository.DeleteMessage(ctx, msgID)
	if err != nil {
		return err
	}
	log.Debug("message deleted")

	return nil
}
