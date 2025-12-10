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
	chats map[uuid.UUID]*models.Chat
}

func NewInMemoryChatRepository() *InMemoryChatRepository {
	repo := InMemoryChatRepository{
		chats: map[uuid.UUID]*models.Chat{},
	}

	return &repo
}

func (r *InMemoryChatRepository) GetChat(id uuid.UUID) (*models.Chat, error) {
	chat, ok := r.chats[id]
	if !ok {
		return nil, errors.New("chat with ID {} does not exist")
	}
	return chat, nil
}

func (r *InMemoryChatRepository) CreateChat(c *models.Chat) error {
	if _, ok := r.chats[c.ID]; ok {
		return errors.New("chat with ID {} already exists")
	}
	r.chats[c.ID] = c
	return nil
}

func (r *InMemoryChatRepository) AddChatUser(id uuid.UUID, userId uuid.UUID) error {
	r.chats[id].UserIDs = append(r.chats[id].UserIDs, userId)
	return nil
}

func (r *InMemoryChatRepository) DeleteChat(id uuid.UUID) error {
	delete(r.chats, id)
	return nil
}
