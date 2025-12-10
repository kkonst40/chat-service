package repositories

import (
	"errors"

	"github.com/google/uuid"
	"github.com/kkonst40/ichat/internal/domain/models"
)

type ChatRepository interface {
	GetChat(id uuid.UUID) (*models.Chat, error)
	CreateChat(c *models.Chat) error
	AddChatUser(id, userId uuid.UUID) error
	DeleteChat(id uuid.UUID) error
}

type InMemoryChatRepository struct {
	Chats map[uuid.UUID]*models.Chat
}

func (r *InMemoryChatRepository) GetChat(id uuid.UUID) (*models.Chat, error) {
	chat, ok := r.Chats[id]
	if !ok {
		return nil, errors.New("chat with ID {} does not exist")
	}
	return chat, nil
}

func (r *InMemoryChatRepository) CreateChat(c *models.Chat) error {
	if _, ok := r.Chats[c.ID]; ok {
		return errors.New("chat with ID {} already exists")
	}
	r.Chats[c.ID] = c
	return nil
}

func (r *InMemoryChatRepository) AddChatUser(id uuid.UUID, userId uuid.UUID) error {
	r.Chats[id].UserIDs = append(r.Chats[id].UserIDs, userId)
	return nil
}

func (r *InMemoryChatRepository) DeleteChat(id uuid.UUID) error {
	delete(r.Chats, id)
	return nil
}
