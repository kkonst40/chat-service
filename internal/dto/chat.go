package dto

import (
	"time"

	"github.com/google/uuid"
)

type GetChatResponse struct {
	ID            uuid.UUID `json:"id"`
	Name          string    `json:"name"`
	LastMessageAt time.Time `json:"lastMessageAt"`
}

type GetChatsResponse struct {
	Chats []GetChatResponse `json:"chats"`
}

type CreateChatRequest struct {
	Name    string      `json:"name" validate:"required"`
	UserIDs []uuid.UUID `json:"userIds" validate:"required"`
}

type UpdateChatNameRequest struct {
	Name string `json:"name" validate:"required"`
}
