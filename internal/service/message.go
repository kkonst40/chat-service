package service

import (
	"fmt"
	"time"

	"github.com/google/uuid"
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

func (s *MessageService) GetChatMessages(chatID uuid.UUID, requesterID uuid.UUID) ([]*model.Message, error) {
	if _, err := s.userService.GetChatUser(chatID, requesterID); err != nil {
		return nil, fmt.Errorf("user is not in the chat")
	}
	if _, err := s.chatService.GetChat(chatID); err != nil {
		return nil, err
	}
	messages, err := s.messageRepository.GetChatMessages(chatID)
	return messages, err
}

func (s *MessageService) CreateMessage(userID, chatID uuid.UUID, text string) (*model.Message, error) {
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

	err = s.messageRepository.CreateMessage(message)
	return message, err
}

func (s *MessageService) UpdateMessage(msgID uuid.UUID, text string, requesterID uuid.UUID) error {
	message, err := s.messageRepository.GetMessage(msgID)
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

	err = s.messageRepository.UpdateMessage(newMessage)
	return err
}

func (s *MessageService) DeleteMessage(msgID uuid.UUID, requesterID uuid.UUID) error {
	message, err := s.messageRepository.GetMessage(msgID)
	if err != nil {
		return err
	}

	if message.UserID == requesterID {
		err = s.messageRepository.DeleteMessage(msgID)
		return err
	}

	sender, err := s.userService.GetChatUser(message.ChatID, message.UserID)
	if err != nil {
		return err
	}
	requester, err := s.userService.GetChatUser(message.ChatID, requesterID)
	if err != nil {
		return err
	}

	if requester.Role <= sender.Role {
		return fmt.Errorf("user has no permission")
	}

	return nil
}
