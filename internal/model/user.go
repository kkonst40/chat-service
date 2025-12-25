package model

import "github.com/google/uuid"

type User struct {
	ID     uuid.UUID `json:"id"`
	ChatID uuid.UUID `json:"chatId"`
	Role   `json:"role"`
}

type Role int

const (
	Common Role = iota
	Admin
)
