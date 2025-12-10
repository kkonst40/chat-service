package services

import (
	"github.com/google/uuid"
	"github.com/kkonst40/ichat/internal/domain/models"
	"github.com/kkonst40/ichat/internal/repositories"
)

type MessageService struct {
	messageRepository repositories.MessageRepository
}

func NewMessageService() *MessageService {
	repo := repositories.NewInMemoryMessageRepository()
	service := MessageService{
		messageRepository: repo,
	}

	return &service
}

func (s *MessageService) GetMessage(id uuid.UUID) (*models.Message, error) {
	message, err := s.messageRepository.GetMessage(id)
	return message, err
}

func (s *MessageService) CreateMessage(userID, chatID uuid.UUID, text string) error {
	newId, err := uuid.NewV7()
	if err != nil {
		return err
	}

	message := &models.Message{
		ID:     newId,
		UserID: userID,
		ChatID: chatID,
		Text:   text,
	}

	err = s.messageRepository.CreateMessage(message)
	return err
}

func (s *MessageService) UpdateMessage(id, userID, chatID uuid.UUID, text string) error {
	message := &models.Message{
		ID:     id,
		UserID: userID,
		ChatID: chatID,
		Text:   text,
	}

	err := s.messageRepository.UpdateMessage(message)
	return err
}

func (s *MessageService) DeleteMessage(id uuid.UUID) error {
	err := s.messageRepository.DeleteMessage(id)
	return err
}
