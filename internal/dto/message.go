package dto

import (
	"time"

	"github.com/google/uuid"
)

type GetMessageResponse struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"userId"`
	ChatID    uuid.UUID `json:"chatId"`
	Text      string    `json:"text"`
	CreatedAt time.Time `json:"createdAt"`
}

type GetMessagesResponse struct {
	Messages []GetMessageResponse `json:"messages"`
}

type AddChatUsersRequest struct {
	UserIDs []uuid.UUID `json:"users" binding:"required"`
}
