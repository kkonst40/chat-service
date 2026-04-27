package memory

import (
	"context"
	"fmt"
	"sort"

	"github.com/google/uuid"
	errs "github.com/kkonst40/chat-service/internal/domain/errors"
	"github.com/kkonst40/chat-service/internal/domain/model"
	"github.com/kkonst40/chat-service/internal/repository"
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

func (r *ChatRepository) GetUserChats(ctx context.Context, userID uuid.UUID, filter model.ChatFilter) ([]model.Chat, error) {
	r.db.mu.RLock()
	defer r.db.mu.RUnlock()

	chats := make([]model.Chat, 0)
	switch filter {
	case model.AllChats:
		for _, user := range r.db.users {
			if user.ID == userID {
				chats = append(chats, *r.db.chats[user.ChatID])
			}
		}
	case model.PersonalChats:
		for _, user := range r.db.users {
			if user.ID == userID && !r.db.chats[user.ChatID].IsGroup {
				chats = append(chats, *r.db.chats[user.ChatID])
			}
		}
	case model.GroupChats:
		for _, user := range r.db.users {
			if user.ID == userID && r.db.chats[user.ChatID].IsGroup {
				chats = append(chats, *r.db.chats[user.ChatID])
			}
		}
	}

	sort.Slice(chats, func(i, j int) bool {
		return chats[i].LastMessageAt.After(chats[j].LastMessageAt)
	})

	return chats, nil
}

func (r *ChatRepository) CreateGroupChat(ctx context.Context, chat *model.Chat, creatorID uuid.UUID, userIDs []uuid.UUID) error {
	r.db.mu.Lock()
	defer r.db.mu.Unlock()

	if _, ok := r.db.chats[chat.ID]; ok {
		return errs.ErrDatabase
	}
	r.db.chats[chat.ID] = chat

	r.db.users[key{ChatID: chat.ID, UserID: creatorID}] = &model.User{
		ID:     creatorID,
		ChatID: chat.ID,
		Role:   model.Owner,
	}

	for _, userID := range userIDs {
		r.db.users[key{ChatID: chat.ID, UserID: userID}] = &model.User{
			ID:     userID,
			ChatID: chat.ID,
			Role:   model.Common,
		}
	}

	return nil
}

func (r *ChatRepository) CreatePersonalChat(ctx context.Context, chat *model.Chat, userID1, userID2 uuid.UUID) error {
	r.db.mu.Lock()
	defer r.db.mu.Unlock()

	if _, ok := r.db.chats[chat.ID]; ok {
		return errs.ErrDatabase
	}
	r.db.chats[chat.ID] = chat

	r.db.users[key{ChatID: chat.ID, UserID: userID1}] = &model.User{
		ID:     userID1,
		ChatID: chat.ID,
		Role:   model.Owner,
	}

	r.db.users[key{ChatID: chat.ID, UserID: userID2}] = &model.User{
		ID:     userID2,
		ChatID: chat.ID,
		Role:   model.Owner,
	}

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

	for k := range r.db.users {
		if k.ChatID == chatID {
			delete(r.db.users, k)
		}
	}

	for id, msg := range r.db.messages {
		if msg.ChatID == chatID {
			delete(r.db.messages, id)
		}
	}

	delete(r.db.chats, chatID)

	return nil
}

func (r *ChatRepository) DeletePersonalChat(ctx context.Context, userID1, userID2 uuid.UUID) error {
	r.db.mu.Lock()
	defer r.db.mu.Unlock()

	var targetChatID uuid.UUID
	found := false

	user1Chats := make(map[uuid.UUID]bool)
	for k := range r.db.users {
		if k.UserID == userID1 {
			user1Chats[k.ChatID] = true
		}
	}

	for k := range r.db.users {
		if k.UserID == userID2 {
			if _, exists := user1Chats[k.ChatID]; exists {
				chat, ok := r.db.chats[k.ChatID]
				if ok && !chat.IsGroup {
					targetChatID = k.ChatID
					found = true
					break
				}
			}
		}
	}

	if !found {
		return fmt.Errorf("private chat not found")
	}

	for k := range r.db.users {
		if k.ChatID == targetChatID {
			delete(r.db.users, k)
		}
	}

	for id, msg := range r.db.messages {
		if msg.ChatID == targetChatID {
			delete(r.db.messages, id)
		}
	}

	delete(r.db.chats, targetChatID)

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
