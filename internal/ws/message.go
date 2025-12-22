package ws

import (
	"github.com/google/uuid"
)

type message struct {
	userID uuid.UUID
	data   []byte
}

type jsonMessage struct {
	userID string `json:"userId"`
	text   string `json:"text"`
}
