package services

import (
	"github.com/google/uuid"
	"github.com/kkonst40/ichat/internal/domain/models"
	"github.com/kkonst40/ichat/internal/repositories"
)

type MesssageService struct {
	MessageRepository repositories.MessageRepository
}

func (s *MesssageService) GetMessage(id uuid.UUID) (*models.Message, error) {
	message, err := s.MessageRepository.GetMessage(id)
	return message, err
}

func (s *MesssageService) CreateMessage(userID, chatID uuid.UUID, text string) error {
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

	err = s.MessageRepository.CreateMessage(message)
	return err
}

func (s *MesssageService) UpdateMessage(id, userID, chatID uuid.UUID, text string) error {
	message := &models.Message{
		ID:     id,
		UserID: userID,
		ChatID: chatID,
		Text:   text,
	}

	err := s.MessageRepository.UpdateMessage(message)
	return err
}

func (s *MesssageService) DeleteMessage(id uuid.UUID) error {
	err := s.MessageRepository.DeleteMessage(id)
	return err
}
