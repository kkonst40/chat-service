package dto

import "github.com/google/uuid"

type GetUserResponse struct {
	ID     uuid.UUID `json:"id"`
	ChatID uuid.UUID `json:"chatId"`
	Role   string    `json:"role"`
}

type GetChatUsersResponse struct {
	Users []GetUserResponse
}

type UpdateChatUserRoleRequest struct {
	Role string `json:"role" binding:"required"`
}
