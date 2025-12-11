package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/kkonst40/ichat/internal/services"
)

type MessageHandler struct {
	messageService *services.MessageService
}

func NewMessageHandler(newMessageService *services.MessageService) *MessageHandler {
	handler := MessageHandler{
		messageService: newMessageService,
	}

	return &handler
}

func (h *MessageHandler) GetChatMessages() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		idParam := ctx.Param("id")
		id, err := uuid.Parse(idParam)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid chat ID format",
			})
			return
		}

		messages, err := h.messageService.GetChatMessages(id)
		if err != nil {
			ctx.JSON(http.StatusNotFound, gin.H{
				"error": "Chat not found",
			})
			return
		}

		ctx.JSON(http.StatusOK, messages)
	}
}

func (h *MessageHandler) SendMessages() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req struct {
			UserID uuid.UUID `json:"user_id" binding:"required"`
			ChatID uuid.UUID `json:"chat_id" binding:"required"`
			Text   string    `json:"text" binding:"required"`
		}

		if err := ctx.ShouldBindJSON(&req); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error":   "Invalid request body",
				"details": err.Error(),
			})
			return
		}

		_, err := h.messageService.CreateMessage(req.UserID, req.ChatID, req.Text)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to send message",
				"details": err.Error(),
			})
			return
		}

		ctx.Status(http.StatusCreated)
	}
}
