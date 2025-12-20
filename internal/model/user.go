package model

import "github.com/google/uuid"

type User struct {
	ID     uuid.UUID
	ChatID uuid.UUID
	Role
}

type Role int

const (
	Common Role = iota
	Admin
)
