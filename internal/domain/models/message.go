package models

import (
	"time"

	"github.com/google/uuid"
)

type Message struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"user_id"`
	ChatID    uuid.UUID `json:"chat_id"`
	Text      string    `json:"text"`
	CreatedAt time.Time `json:"created_at"`
}
