package model

import (
	"time"

	"github.com/google/uuid"
)

type Message struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"userId"`
	ChatID    uuid.UUID `json:"chatId"`
	Text      string    `json:"text"`
	CreatedAt time.Time `json:"createdAt"`
}
