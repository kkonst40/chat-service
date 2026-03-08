package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/kkonst40/ichat/internal/domain/model"
)

type UserRepository interface {
	GetChatUser(ctx context.Context, chatID, userID uuid.UUID) (*model.User, error)
	GetChatUsers(ctx context.Context, chatID uuid.UUID) ([]model.User, error)
	GetChatUserIDs(ctx context.Context, chatID uuid.UUID) ([]uuid.UUID, error)
	GetPersonalChatsInterlocutors(ctx context.Context, userID uuid.UUID) (map[uuid.UUID]uuid.UUID, error)
	AddChatUsers(ctx context.Context, chatID uuid.UUID, userIDs []uuid.UUID) ([]uuid.UUID, error)
	DeleteChatUser(ctx context.Context, chatID uuid.UUID, userID uuid.UUID) error
	UpdateUserRole(ctx context.Context, chatID, userID uuid.UUID, newRole model.Role) error
	UserInChat(ctx context.Context, chatID, userID uuid.UUID) (bool, error)
}
