package models

import "github.com/google/uuid"

type Message struct {
	ID     uuid.UUID
	UserID uuid.UUID
	ChatID uuid.UUID
	Text   string
}
