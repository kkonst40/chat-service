package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/kkonst40/ichat/internal/model"
)

type MessageRepository interface {
	GetMessage(ctx context.Context, msgID uuid.UUID) (*model.Message, error)
	GetChatMessages(ctx context.Context, chatID uuid.UUID) ([]model.Message, error)
	CreateMessage(ctx context.Context, msg *model.Message) error
	UpdateMessage(ctx context.Context, msg *model.Message) error
	DeleteMessage(ctx context.Context, msgID uuid.UUID) error
}
