package model

import "github.com/google/uuid"

type User struct {
	ID     uuid.UUID
	ChatID uuid.UUID
	Role
}

type Role string

const (
	Common Role = "common"
	Admin  Role = "admin"
	Owner  Role = "owner"
)

var UserPriority map[Role]int = map[Role]int{
	Common: 0,
	Admin:  1,
	Owner:  2,
}
