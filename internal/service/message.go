package service

import (
	"time"

	"github.com/google/uuid"
	"github.com/kkonst40/ichat/internal/model"
	"github.com/kkonst40/ichat/internal/repository"
)

type MessageService struct {
	messageRepository repository.MessageRepository
	chatService       *ChatService
}

func NewMessageService(
	newMessageRepository repository.MessageRepository,
	newChatService *ChatService,
) *MessageService {
	service := MessageService{
		messageRepository: newMessageRepository,
		chatService:       newChatService,
	}

	return &service
}

func (s *MessageService) GetMessage(id uuid.UUID) (*model.Message, error) {
	message, err := s.messageRepository.GetMessage(id)
	return message, err
}

func (s *MessageService) GetChatMessages(chatId uuid.UUID) ([]*model.Message, error) {
	if _, err := s.chatService.GetChat(chatId); err != nil {
		return nil, err
	}
	messages, err := s.messageRepository.GetChatMessages(chatId)
	return messages, err
}

func (s *MessageService) CreateMessage(userID, chatID uuid.UUID, text string) (*model.Message, error) {
	newId, err := uuid.NewV7()
	if err != nil {
		return nil, err
	}

	message := &model.Message{
		ID:        newId,
		UserID:    userID,
		ChatID:    chatID,
		Text:      text,
		CreatedAt: time.Now(),
	}

	err = s.messageRepository.CreateMessage(message)
	return message, err
}

func (s *MessageService) UpdateMessage(id uuid.UUID, text string) error {
	message, err := s.messageRepository.GetMessage(id)
	if err != nil {
		return err
	}

	newMessage := &model.Message{
		ID:        id,
		UserID:    message.UserID,
		ChatID:    message.ChatID,
		Text:      text,
		CreatedAt: message.CreatedAt,
	}

	err = s.messageRepository.UpdateMessage(newMessage)
	return err
}

func (s *MessageService) DeleteMessage(id uuid.UUID) error {
	err := s.messageRepository.DeleteMessage(id)
	return err
}
