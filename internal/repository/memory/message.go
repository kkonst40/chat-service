package memory

import (
	"context"
	"sort"

	"github.com/google/uuid"
	errs "github.com/kkonst40/ichat/internal/errors"
	"github.com/kkonst40/ichat/internal/model"
	"github.com/kkonst40/ichat/internal/repository"
)

type MessageRepository struct {
	db *MemoryDB
}

func NewMessageRepository(db *MemoryDB) *MessageRepository {
	return &MessageRepository{
		db: db,
	}
}

func (r *MessageRepository) GetMessage(ctx context.Context, msgID uuid.UUID) (*model.Message, error) {
	r.db.mu.Lock()
	defer r.db.mu.Unlock()

	message, ok := r.db.messages[msgID]
	if !ok {
		return nil, errs.ErrMsgNotFound
	}

	return message, nil
}

func (r *MessageRepository) GetChatMessages(ctx context.Context, chatID uuid.UUID, from, count int64) ([]model.Message, error) {
	r.db.mu.Lock()
	defer r.db.mu.Unlock()

	messages := make([]model.Message, 0)
	for _, v := range r.db.messages {
		if v.ChatID == chatID {
			messages = append(messages, *v)
		}
	}

	sort.Slice(messages, func(i, j int) bool {
		return messages[i].CreatedAt.After(messages[j].CreatedAt)
	})

	return messages[from : from+count], nil
}

func (r *MessageRepository) CreateMessage(ctx context.Context, msg *model.Message) error {
	r.db.mu.Lock()
	defer r.db.mu.Unlock()

	if _, ok := r.db.messages[msg.ID]; ok {
		return errs.ErrDatabase
	}
	r.db.messages[msg.ID] = msg

	return nil
}

func (r *MessageRepository) UpdateMessage(ctx context.Context, msg *model.Message) error {
	r.db.mu.Lock()
	defer r.db.mu.Unlock()

	if _, ok := r.db.messages[msg.ID]; !ok {
		return errs.ErrMsgNotFound
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

var _ repository.MessageRepository = (*MessageRepository)(nil)
