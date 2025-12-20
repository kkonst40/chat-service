package repository

import (
	"fmt"
	"sync"

	"github.com/google/uuid"
	"github.com/kkonst40/ichat/internal/model"
)

type MessageRepository interface {
	GetMessage(id uuid.UUID) (*model.Message, error)
	GetChatMessages(chatId uuid.UUID) ([]*model.Message, error)
	CreateMessage(m *model.Message) error
	UpdateMessage(m *model.Message) error
	DeleteMessage(id uuid.UUID) error
}

type InMemoryMessageRepository struct {
	messages map[uuid.UUID]*model.Message
	mu       sync.Mutex
}

func NewInMemoryMessageRepository() *InMemoryMessageRepository {
	repo := InMemoryMessageRepository{
		messages: make(map[uuid.UUID]*model.Message),
		mu:       sync.Mutex{},
	}

	return &repo
}

func (r *InMemoryMessageRepository) GetMessage(id uuid.UUID) (*model.Message, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	message, ok := r.messages[id]
	if !ok {
		return nil, fmt.Errorf("message with ID %s does not exist", id)
	}
	return message, nil
}

func (r *InMemoryMessageRepository) GetChatMessages(chatId uuid.UUID) ([]*model.Message, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	messages := make([]*model.Message, 0)
	for _, v := range r.messages {
		if v.ChatID == chatId {
			messages = append(messages, v)
		}
	}
	return messages, nil
}

func (r *InMemoryMessageRepository) CreateMessage(m *model.Message) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.messages[m.ID]; ok {
		return fmt.Errorf("message with ID %s already exists", m.ID)
	}
	r.messages[m.ID] = m
	return nil
}

func (r *InMemoryMessageRepository) UpdateMessage(m *model.Message) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.messages[m.ID]; !ok {
		return fmt.Errorf("message with ID %s does not exist", m.ID)
	}
	r.messages[m.ID] = m
	return nil
}

func (r *InMemoryMessageRepository) DeleteMessage(id uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.messages, id)
	return nil
}
