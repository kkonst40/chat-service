package models

import "github.com/google/uuid"

type Chat struct {
	ID      uuid.UUID
	UserIDs []uuid.UUID
}
