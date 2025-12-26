package repository

import (
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
