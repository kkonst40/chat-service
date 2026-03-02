package event

import (
	"github.com/google/uuid"
	"github.com/kkonst40/ichat/internal/domain/model"
)

type EventType string

const (
	CreateMsg EventType = "CREATE_MESSAGE"
	UpdateMsg EventType = "UPDATE_MESSAGE"
	DeleteMsg EventType = "DELETE_MESSAGE"

	CreateChat EventType = "CREATE_CHAT"
	UpdateChat EventType = "UPDATE_CHAT"
	DeleteChat EventType = "DELETE_CHAT"

	CreateChatUser EventType = "CREATE_CHATUSER"
	UpdateChatUser EventType = "UPDATE_CHATUSER"
	DeleteChatUser EventType = "DELETE_CHATUSER"
)

type Event struct {
	Type    EventType `json:"type"`
	ChatID  uuid.UUID `json:"chat_id"`
	Payload any       `json:"payload"`
}

type CreateMsgEvent struct {
	MsgID  uuid.UUID `json:"msg_id"`
	UserID uuid.UUID `json:"user_id"`
	Text   string    `json:"text"`
}

type UpdateMsgEvent struct {
	MsgID uuid.UUID `json:"msg_id"`
	Text  string    `json:"text"`
}

type DeleteMsgEvent struct {
	MsgID uuid.UUID `json:"msg_id"`
}

type CreateUserEvent struct {
	UserID uuid.UUID `json:"user_id"`
}

type UpdateUserEvent struct {
	UserID uuid.UUID  `json:"user_id"`
	Role   model.Role `json:"role"`
}

type CreateChatEvent struct {
	Name string `json:"name"`
}

type UpdateChatEvent struct {
	Name string `json:"name"`
}

type DeleteChatEvent struct {
}
