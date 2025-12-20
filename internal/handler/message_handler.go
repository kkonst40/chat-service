package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
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
		id, err := uuid.Parse(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid chat ID format",
			})
			return
		}

		messages, err := h.messageService.GetChatMessages(id)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Chat not found",
			})
			return
		}

		c.JSON(http.StatusOK, messages)
	}
}

func (h *MessageHandler) SendMessages() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			ChatID uuid.UUID `json:"chatId" binding:"required"`
			Text   string    `json:"text" binding:"required"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Invalid request body",
				"details": err.Error(),
			})
			return
		}

		userId, err := uuid.Parse(c.GetString("userID"))
		if err != nil {
			c.Status(http.StatusInternalServerError)
			return
		}

		message, err := h.messageService.CreateMessage(userId, req.ChatID, req.Text)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to send message",
				"details": err.Error(),
			})
			return
		}

		c.JSON(http.StatusCreated, message)
	}
}
