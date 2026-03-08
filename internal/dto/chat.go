package dto

import (
	"time"

	"github.com/google/uuid"
)

type GetChatResponse struct {
	ID            uuid.UUID `json:"id"`
	Name          string    `json:"name"`
	IsGroup       bool      `json:"isGroup"`
	LastMessageAt time.Time `json:"lastMessageAt"`
}

type GetChatsResponse struct {
	Chats []GetChatResponse `json:"chats"`
}

type CreateGroupChatRequest struct {
	Name      string   `json:"name" validate:"required"`
	UserNames []string `json:"userNames" validate:"required"`
}

type CreatePersonalChatRequest struct {
	UserName string `json:"userName" validate:"required"`
}

type UpdateChatNameRequest struct {
	Name string `json:"name" validate:"required"`
}
