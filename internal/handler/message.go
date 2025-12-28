package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/kkonst40/ichat/internal/dto"
	"github.com/kkonst40/ichat/internal/service"
)

type MessageHandler struct {
	messageService *service.MessageService
}

func NewMessageHandler(newMessageService *service.MessageService) *MessageHandler {
	handler := MessageHandler{
		messageService: newMessageService,
	}

	return &handler
}

func (h *MessageHandler) GetChatMessages() gin.HandlerFunc {
	return func(c *gin.Context) {
		requesterID := uuid.MustParse(c.GetString("requesterID"))
		ctx := c.Request.Context()

		chatID, err := uuid.Parse(c.Param("chatId"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid chat ID format",
			})
			return
		}

		messages, err := h.messageService.GetChatMessages(ctx, chatID, requesterID)
		if err != nil {
			//
			return
		}

		resp := dto.GetMessagesResponse{
			Messages: make([]dto.GetMessageResponse, 0, len(messages)),
		}

		for _, message := range messages {
			resp.Messages = append(resp.Messages, dto.GetMessageResponse{
				ID:        message.ID,
				UserID:    message.UserID,
				ChatID:    message.ChatID,
				Text:      message.Text,
				CreatedAt: message.CreatedAt,
			})
		}

		c.JSON(http.StatusOK, resp)
	}
}
