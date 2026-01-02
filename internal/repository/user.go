package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/kkonst40/ichat/internal/model"
)

type UserRepository interface {
	GetChatUser(ctx context.Context, chatID, userID uuid.UUID) (*model.User, error)
	GetChatUsers(ctx context.Context, chatID uuid.UUID) ([]*model.User, error)
	AddChatUsers(ctx context.Context, chatID uuid.UUID, userIDs []uuid.UUID) error
	DeleteChatUser(ctx context.Context, chatID uuid.UUID, userID uuid.UUID) error
	UpdateUserRole(ctx context.Context, chatID, userID uuid.UUID, newRole model.Role) error
	IsUserInChat(ctx context.Context, chatID, userID uuid.UUID) (bool, error)
}
