package ws

import (
	"github.com/google/uuid"
)

type message struct {
	userID uuid.UUID
	data   []byte
}

type jsonMessage struct {
	UserID string `json:"userId"`
	Text   string `json:"text"`
}
