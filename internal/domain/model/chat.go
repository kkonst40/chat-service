package model

import (
	"time"

	"github.com/google/uuid"
)

type Chat struct {
	ID            uuid.UUID
	Name          string
	IsGroup       bool
	LastMessageAt time.Time
}

type ChatFilter string

const (
	AllChats      = ChatFilter("all")
	PersonalChats = ChatFilter("personal")
	GroupChats    = ChatFilter("group")
)
