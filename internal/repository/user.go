package repository

import (
	"github.com/google/uuid"
	"github.com/kkonst40/ichat/internal/model"
)

type UserRepository interface {
	GetChatUser(chatID, userID uuid.UUID) (*model.User, error)
	GetChatUsers(chatID uuid.UUID) ([]*model.User, error)
	GetUserChatIds(userID uuid.UUID) ([]uuid.UUID, error)
	AddChatUsers(chatID uuid.UUID, userIDs []uuid.UUID) error
	DeleteChatUser(chatID uuid.UUID, userID uuid.UUID) error
	SetUserRole(chatID, userID uuid.UUID, newRole model.Role) error
}
