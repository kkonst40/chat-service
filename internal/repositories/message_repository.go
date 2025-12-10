package repositories

import (
	"errors"

	"github.com/google/uuid"
	"github.com/kkonst40/ichat/internal/domain/models"
)

type MessageRepository interface {
	GetMessage(id uuid.UUID) (*models.Message, error)
	CreateMessage(m *models.Message) error
	UpdateMessage(m *models.Message) error
	DeleteMessage(id uuid.UUID) error
}

type InMemoryMessageRepository struct {
	Messages map[uuid.UUID]*models.Message
}

func (r *InMemoryMessageRepository) GetMessage(id uuid.UUID) (*models.Message, error) {
	message, ok := r.Messages[id]
	if !ok {
		return nil, errors.New("message with ID {} does not exist")
	}
	return message, nil
}

func (r *InMemoryMessageRepository) CreateMessage(m *models.Message) error {
	if _, ok := r.Messages[m.ID]; ok {
		return errors.New("message with ID {} already exists")
	}
	r.Messages[m.ID] = m
	return nil
}

func (r *InMemoryMessageRepository) UpdateMessage(m *models.Message) error {
	if _, ok := r.Messages[m.ID]; !ok {
		return errors.New("message with ID {} does not exists")
	}
	r.Messages[m.ID] = m
	return nil
}

func (r *InMemoryMessageRepository) DeleteMessage(id uuid.UUID) error {
	delete(r.Messages, id)
	return nil
}
