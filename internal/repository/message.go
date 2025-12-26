package repository

import (
	"github.com/google/uuid"
	"github.com/kkonst40/ichat/internal/model"
)

type MessageRepository interface {
	GetMessage(msgID uuid.UUID) (*model.Message, error)
	GetChatMessages(chatID uuid.UUID) ([]*model.Message, error)
	CreateMessage(msg *model.Message) error
	UpdateMessage(msg *model.Message) error
	DeleteMessage(msgID uuid.UUID) error
}
