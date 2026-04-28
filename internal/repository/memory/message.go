package memory

import (
	"context"
	"sort"
	"time"

	"github.com/google/uuid"
	"github.com/kkonst40/chat-service/internal/domain/model"
	"github.com/kkonst40/chat-service/internal/repository"
)

type MessageRepository struct {
	db *MemoryDB
}

func NewMessageRepository(db *MemoryDB) *MessageRepository {
	return &MessageRepository{
		db: db,
	}
}

func (r *MessageRepository) GetMessage(ctx context.Context, msgID uuid.UUID) (model.Message, error) {
	r.db.mu.Lock()
	defer r.db.mu.Unlock()

	message, ok := r.db.messages[msgID]
	if !ok {
		return model.Message{}, repository.ErrNotFound
	}

	return *message, nil
}

func (r *MessageRepository) GetChatMessages(ctx context.Context, chatID uuid.UUID, from uuid.UUID, count int64) ([]model.Message, error) {
	r.db.mu.Lock()
	defer r.db.mu.Unlock()

	messages := make([]model.Message, 0)
	fromTime := r.db.messages[from].CreatedAt
	for _, v := range r.db.messages {
		if v.ChatID == chatID && v.CreatedAt.Before(fromTime) {
			messages = append(messages, *v)
		}
	}

	sort.Slice(messages[:count], func(i, j int) bool {
		return messages[i].CreatedAt.After(messages[j].CreatedAt)
	})

	return messages[:count], nil
}

func (r *MessageRepository) CreateMessage(ctx context.Context, msg *model.Message) error {
	r.db.mu.Lock()
	defer r.db.mu.Unlock()

	if _, ok := r.db.messages[msg.ID]; ok {
		return repository.ErrDatabase
	}
	r.db.messages[msg.ID] = msg
	r.db.chats[msg.ChatID].LastMessageAt = time.Now()

	return nil
}

func (r *MessageRepository) UpdateMessage(ctx context.Context, msg *model.Message) error {
	r.db.mu.Lock()
	defer r.db.mu.Unlock()

	if _, ok := r.db.messages[msg.ID]; !ok {
		return repository.ErrNotFound
	}
	r.db.messages[msg.ID] = msg

	return nil
}

func (r *MessageRepository) DeleteMessage(ctx context.Context, msgID uuid.UUID) error {
	r.db.mu.Lock()
	defer r.db.mu.Unlock()

	delete(r.db.messages, msgID)

	return nil
}
