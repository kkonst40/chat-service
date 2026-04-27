package dto

import (
	"github.com/google/uuid"
	"github.com/kkonst40/chat-service/internal/domain/model"
)

type GetUserResponse struct {
	ID     uuid.UUID `json:"id"`
	ChatID uuid.UUID `json:"chatId"`
	Role   string    `json:"role"`
}

type GetChatUsersResponse struct {
	Users []GetUserResponse `json:"users"`
}

type UpdateChatUserRoleRequest struct {
	Role model.Role `json:"role" validate:"required"`
}

type AddChatUsersRequest struct {
	UserNames []string `json:"userNames" validate:"required"`
}
