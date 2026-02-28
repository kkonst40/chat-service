package model

import (
	"time"

	"github.com/google/uuid"
)

type Chat struct {
	ID            uuid.UUID `json:"id"`
	Name          string    `json:"name"`
	LastMessageAt time.Time `json:"lastMessageAt"`
}
