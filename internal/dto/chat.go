package dto

import "github.com/google/uuid"

type GetChatResponse struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
}

type GetChatsResponse struct {
	Chats []GetChatResponse `json:"chats"`
}

type CreateChatRequest struct {
	Name string `json:"name" binding:"required"`
}

type UpdateChatNameRequest struct {
	Name string `json:"name" binding:"required"`
}

type AddChatUsersRequest struct {
	Users []uuid.UUID `json:"users" binding:"required"`
}
