package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/kkonst40/chat-service/internal/domain/model"
)

type ChatRepository interface {
	GetChat(ctx context.Context, chatID uuid.UUID) (*model.Chat, error)
	GetUserChats(ctx context.Context, userID uuid.UUID, filter model.ChatFilter) ([]model.Chat, error)
	CreateGroupChat(ctx context.Context, chat *model.Chat, creatorID uuid.UUID, userIDs []uuid.UUID) error
	CreatePersonalChat(ctx context.Context, chat *model.Chat, userID1, userID2 uuid.UUID) error
	UpdateChatName(ctx context.Context, chatID uuid.UUID, name string) error
	DeleteChat(ctx context.Context, chatID uuid.UUID) error
	DeletePersonalChat(ctx context.Context, userID1, userID2 uuid.UUID) error
	ChatExists(ctx context.Context, chatID uuid.UUID) (bool, error)
}
