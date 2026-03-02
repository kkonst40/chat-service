package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/kkonst40/ichat/internal/domain/model"
)

type ChatRepository interface {
	GetChat(ctx context.Context, chatID uuid.UUID) (*model.Chat, error)
	GetUserChats(ctx context.Context, userID uuid.UUID) ([]model.Chat, error)
	CreateChat(ctx context.Context, chat *model.Chat, creatorID uuid.UUID) error
	UpdateChatName(ctx context.Context, chatID uuid.UUID, name string) error
	DeleteChat(ctx context.Context, chatID uuid.UUID) error
	ChatExists(ctx context.Context, chatID uuid.UUID) (bool, error)
}
