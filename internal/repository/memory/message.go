package memory

import (
	"context"
	"fmt"
	"sync"

	"github.com/google/uuid"
	"github.com/kkonst40/ichat/internal/model"
	"github.com/kkonst40/ichat/internal/repository"
)

type MessageRepository struct {
	messages map[uuid.UUID]*model.Message
	mu       sync.Mutex
}

func NewMessageRepository() *MessageRepository {
	repo := MessageRepository{
		messages: make(map[uuid.UUID]*model.Message),
		mu:       sync.Mutex{},
	}

	return &repo
}

func (r *MessageRepository) GetMessage(ctx context.Context, msgID uuid.UUID) (*model.Message, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	message, ok := r.messages[msgID]
	if !ok {
		return nil, fmt.Errorf("message with ID %s does not exist", msgID)
	}
	return message, nil
}

func (r *MessageRepository) GetChatMessages(ctx context.Context, chatID uuid.UUID) ([]*model.Message, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	messages := make([]*model.Message, 0)
	for _, v := range r.messages {
		if v.ChatID == chatID {
			messages = append(messages, v)
		}
	}
	return messages, nil
}

func (r *MessageRepository) CreateMessage(ctx context.Context, msg *model.Message) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.messages[msg.ID]; ok {
		return fmt.Errorf("message with ID %s already exists", msg.ID)
	}
	r.messages[msg.ID] = msg
	return nil
}

func (r *MessageRepository) UpdateMessage(ctx context.Context, msg *model.Message) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.messages[msg.ID]; !ok {
		return fmt.Errorf("message with ID %s does not exist", msg.ID)
	}
	r.messages[msg.ID] = msg
	return nil
}

func (r *MessageRepository) DeleteMessage(ctx context.Context, msgID uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.messages, msgID)
	return nil
}

var _ repository.MessageRepository = (*MessageRepository)(nil)
