package model

import (
	"time"

	"github.com/google/uuid"
)

type Chat struct {
	ID            uuid.UUID
	Name          string
	LastMessageAt time.Time
}
