package ws

import (
	"github.com/google/uuid"
)

type ActionType string

const (
	ActionCreate ActionType = "CREATE"
	ActionUpdate ActionType = "UPDATE"
	ActionDelete ActionType = "DELETE"
)

type roomEvent struct {
	Type   ActionType `json:"type"`
	Text   string     `json:"text"`
	MsgID  uuid.UUID  `json:"messageId"`
	UserID uuid.UUID  `json:"userId"`
}
