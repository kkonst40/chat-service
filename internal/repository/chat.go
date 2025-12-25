package repository

import (
	"fmt"
	"sync"

	"github.com/google/uuid"
	"github.com/kkonst40/ichat/internal/model"
)

type ChatRepository interface {
	GetChat(chatID uuid.UUID) (*model.Chat, error)
	GetChats(chatIDs []uuid.UUID) ([]*model.Chat, error)
	CreateChat(chat *model.Chat) error
	UpdateChatName(chatID uuid.UUID, name string) error
	DeleteChat(chatID uuid.UUID) error
}

type InMemoryChatRepository struct {
	chats map[uuid.UUID]*model.Chat
	mu    sync.RWMutex
}

func NewInMemoryChatRepository() *InMemoryChatRepository {
	repo := InMemoryChatRepository{
		chats: make(map[uuid.UUID]*model.Chat),
		mu:    sync.RWMutex{},
	}

	id0 := uuid.MustParse("018f95a6-8d27-7e03-822c-6a81ce0d1f4b")
	id1 := uuid.MustParse("018f95a6-8d28-7b41-8d9f-12e8b341a6c5")
	id2 := uuid.MustParse("018f95a6-8d28-7c8a-9123-4f67d890a12e")
	id3 := uuid.MustParse("018f95a6-8d28-7de9-b8a4-5c32f1e09d76")
	id4 := uuid.MustParse("018f95a6-8d29-7023-84d1-9a0b2c3e4f5a")
	id5 := uuid.MustParse("018f95a6-8d29-71bc-9a8b-76e54d32c10f")
	id6 := uuid.MustParse("018f95a6-8d2a-72f4-a123-b456c789d0e1")
	id7 := uuid.MustParse("018f95a6-8d2a-743d-b987-6543210fedcb")
	id8 := uuid.MustParse("018f95a6-8d2b-7586-c246-8ace13579bdf")
	id9 := uuid.MustParse("018f95a6-8d2b-76cf-d369-147f258b036a")

	repo.chats[id0] = &model.Chat{ID: id0, Name: "Chat0"}
	repo.chats[id1] = &model.Chat{ID: id1, Name: "Chat1"}
	repo.chats[id2] = &model.Chat{ID: id2, Name: "Chat2"}
	repo.chats[id3] = &model.Chat{ID: id3, Name: "Chat3"}
	repo.chats[id4] = &model.Chat{ID: id4, Name: "Chat4"}
	repo.chats[id5] = &model.Chat{ID: id5, Name: "Chat5"}
	repo.chats[id6] = &model.Chat{ID: id6, Name: "Chat6"}
	repo.chats[id7] = &model.Chat{ID: id7, Name: "Chat7"}
	repo.chats[id8] = &model.Chat{ID: id8, Name: "Chat8"}
	repo.chats[id9] = &model.Chat{ID: id9, Name: "Chat9"}

	return &repo
}

func (r *InMemoryChatRepository) GetChat(chatID uuid.UUID) (*model.Chat, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	chat, ok := r.chats[chatID]
	if !ok {
		return nil, fmt.Errorf("chat with ID %s does not exist", chatID)
	}
	return chat, nil
}

func (r *InMemoryChatRepository) GetChats(chatIDs []uuid.UUID) ([]*model.Chat, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	chats := make([]*model.Chat, 0, len(chatIDs))
	for _, ID := range chatIDs {
		chats = append(chats, r.chats[ID])
	}
	return chats, nil
}

func (r *InMemoryChatRepository) CreateChat(chat *model.Chat) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.chats[chat.ID]; ok {
		return fmt.Errorf("chat with ID %s already exists", chat.ID)
	}
	r.chats[chat.ID] = chat
	return nil
}

func (r *InMemoryChatRepository) UpdateChatName(chatID uuid.UUID, name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.chats[chatID]; !ok {
		return fmt.Errorf("chat with ID %s does not exist", chatID)
	}
	r.chats[chatID].Name = name
	return nil
}

func (r *InMemoryChatRepository) DeleteChat(chatID uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.chats, chatID)
	return nil
}
