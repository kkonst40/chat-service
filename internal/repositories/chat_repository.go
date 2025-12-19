package repositories

import (
	"fmt"
	"slices"
	"sync"

	"github.com/google/uuid"
	"github.com/kkonst40/ichat/internal/models"
)

type ChatRepository interface {
	GetChat(id uuid.UUID) (*models.Chat, error)
	GetChats(userId uuid.UUID) ([]*models.Chat, error)
	CreateChat(c *models.Chat) error
	UpdateChatName(id uuid.UUID, name string) error
	AddChatUser(id, userId uuid.UUID) error
	DeleteChat(id uuid.UUID) error
}

type InMemoryChatRepository struct {
	chats map[uuid.UUID]*models.Chat
	mu    sync.RWMutex
}

func NewInMemoryChatRepository() *InMemoryChatRepository {
	repo := InMemoryChatRepository{
		chats: make(map[uuid.UUID]*models.Chat),
		mu:    sync.RWMutex{},
	}

	return &repo
}

func (r *InMemoryChatRepository) GetChat(id uuid.UUID) (*models.Chat, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	chat, ok := r.chats[id]
	if !ok {
		return nil, fmt.Errorf("chat with ID %s does not exist", id)
	}
	return chat, nil
}

func (r *InMemoryChatRepository) GetChats(userId uuid.UUID) ([]*models.Chat, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	chats := make([]*models.Chat, 0)
	for _, v := range r.chats {
		if slices.Contains(v.UserIDs, userId) {
			chats = append(chats, v)
		}
	}
	return chats, nil
}

func (r *InMemoryChatRepository) CreateChat(c *models.Chat) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.chats[c.ID]; ok {
		return fmt.Errorf("chat with ID %s already exists", c.ID)
	}
	r.chats[c.ID] = c
	return nil
}

func (r *InMemoryChatRepository) UpdateChatName(id uuid.UUID, name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.chats[id]; !ok {
		return fmt.Errorf("chat with ID %s does not exist", id)
	}
	r.chats[id].Name = name
	return nil
}

func (r *InMemoryChatRepository) AddChatUser(id uuid.UUID, userId uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.chats[id]; !ok {
		return fmt.Errorf("chat with ID %s does not exist", id)
	}
	r.chats[id].UserIDs = append(r.chats[id].UserIDs, userId)
	return nil
}

func (r *InMemoryChatRepository) DeleteChat(id uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.chats, id)
	return nil
}
