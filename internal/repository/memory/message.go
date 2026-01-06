package memory

import (
	"context"
	"fmt"

	"github.com/google/uuid"
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
		return nil, fmt.Errorf("message with ID %s does not exist", msgID)
	}

	return message, nil
}

func (r *MessageRepository) GetChatMessages(ctx context.Context, chatID uuid.UUID) ([]model.Message, error) {
	r.db.mu.Lock()
	defer r.db.mu.Unlock()

	messages := make([]model.Message, 0)
	for _, v := range r.db.messages {
		if v.ChatID == chatID {
			messages = append(messages, *v)
		}
	}

	return messages, nil
}

func (r *MessageRepository) CreateMessage(ctx context.Context, msg *model.Message) error {
	r.db.mu.Lock()
	defer r.db.mu.Unlock()

	if _, ok := r.db.messages[msg.ID]; ok {
		return fmt.Errorf("message with ID %s already exists", msg.ID)
	}
	r.db.messages[msg.ID] = msg

	return nil
}

func (r *MessageRepository) UpdateMessage(ctx context.Context, msg *model.Message) error {
	r.db.mu.Lock()
	defer r.db.mu.Unlock()

	if _, ok := r.db.messages[msg.ID]; !ok {
		return fmt.Errorf("message with ID %s does not exist", msg.ID)
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
