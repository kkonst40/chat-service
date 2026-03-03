package model

import (
	"time"

	"github.com/google/uuid"
)

type Message struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	ChatID    uuid.UUID
	Text      string
	CreatedAt time.Time
}
