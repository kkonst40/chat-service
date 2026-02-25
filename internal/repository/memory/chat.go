package memory

import (
	"context"

	"github.com/google/uuid"
	errs "github.com/kkonst40/ichat/internal/errors"
	"github.com/kkonst40/ichat/internal/model"
	"github.com/kkonst40/ichat/internal/repository"
)

type ChatRepository struct {
	db *MemoryDB
}

func NewChatRepository(db *MemoryDB) *ChatRepository {
	return &ChatRepository{
		db: db,
	}
}

func (r *ChatRepository) GetChat(ctx context.Context, chatID uuid.UUID) (*model.Chat, error) {
	r.db.mu.RLock()
	defer r.db.mu.RUnlock()

	chat, ok := r.db.chats[chatID]
	if !ok {
		return nil, errs.ErrChatNotFound
	}

	return chat, nil
}

func (r *ChatRepository) GetUserChats(ctx context.Context, userID uuid.UUID) ([]model.Chat, error) {
	r.db.mu.RLock()
	defer r.db.mu.RUnlock()

	chats := make([]model.Chat, 0)
	for _, user := range r.db.users {
		if user.ID == userID {
			chats = append(chats, *r.db.chats[user.ChatID])
		}
	}

	return chats, nil
}

func (r *ChatRepository) CreateChat(ctx context.Context, chat *model.Chat, creatorID uuid.UUID) error {
	r.db.mu.Lock()
	defer r.db.mu.Unlock()

	if _, ok := r.db.chats[chat.ID]; ok {
		return errs.ErrDatabase
	}
	r.db.chats[chat.ID] = chat

	return nil
}

func (r *ChatRepository) UpdateChatName(ctx context.Context, chatID uuid.UUID, name string) error {
	r.db.mu.Lock()
	defer r.db.mu.Unlock()

	if _, ok := r.db.chats[chatID]; !ok {
		return errs.ErrChatNotFound
	}
	r.db.chats[chatID].Name = name

	return nil
}

func (r *ChatRepository) DeleteChat(ctx context.Context, chatID uuid.UUID) error {
	r.db.mu.Lock()
	defer r.db.mu.Unlock()

	delete(r.db.chats, chatID)

	return nil
}

func (r *ChatRepository) ChatExists(ctx context.Context, chatID uuid.UUID) (bool, error) {
	r.db.mu.RLock()
	defer r.db.mu.RUnlock()

	if _, ok := r.db.chats[chatID]; ok {
		return true, nil
	}

	return false, nil
}

var _ repository.ChatRepository = (*ChatRepository)(nil)
