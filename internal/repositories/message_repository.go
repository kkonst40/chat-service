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
	messages map[uuid.UUID]*models.Message
}

func NewInMemoryMessageRepository() *InMemoryMessageRepository {
	repo := InMemoryMessageRepository{
		messages: map[uuid.UUID]*models.Message{},
	}

	return &repo
}

func (r *InMemoryMessageRepository) GetMessage(id uuid.UUID) (*models.Message, error) {
	message, ok := r.messages[id]
	if !ok {
		return nil, errors.New("message with ID {} does not exist")
	}
	return message, nil
}

func (r *InMemoryMessageRepository) CreateMessage(m *models.Message) error {
	if _, ok := r.messages[m.ID]; ok {
		return errors.New("message with ID {} already exists")
	}
	r.messages[m.ID] = m
	return nil
}

func (r *InMemoryMessageRepository) UpdateMessage(m *models.Message) error {
	if _, ok := r.messages[m.ID]; !ok {
		return errors.New("message with ID {} does not exists")
	}
	r.messages[m.ID] = m
	return nil
}

func (r *InMemoryMessageRepository) DeleteMessage(id uuid.UUID) error {
	delete(r.messages, id)
	return nil
}
