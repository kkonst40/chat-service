package models

import "github.com/google/uuid"

type Chat struct {
	ID      uuid.UUID   `json:"id"`
	Name    string      `json:"name"`
	UserIDs []uuid.UUID `json:"user_ids"`
}
