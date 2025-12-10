package handlers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/kkonst40/ichat/internal/services"
)

type ChatHandler struct {
	chatService *services.ChatService
}

func NewChatHandler() *ChatHandler {
	service := services.NewChatService()
	handler := ChatHandler{
		chatService: service,
	}

	return &handler
}

func (h *ChatHandler) GetChat() gin.HandlerFunc {
	return func(c *gin.Context) {
		idParam := c.Param("id")
		id, err := uuid.Parse(idParam)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid chat ID format",
			})
			return
		}

		chat, err := h.chatService.GetChat(id)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Chat not found",
			})
			return
		}

		c.JSON(http.StatusOK, chat)
	}
}

func (h *ChatHandler) GetChats() gin.HandlerFunc {
	return func(c *gin.Context) {
		chats, _ := h.chatService.GetChats()
		c.JSON(http.StatusOK, chats)
	}
}

func (h *ChatHandler) CreateChat() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Name string `json:"name" binding:"required"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Invalid request body",
				"details": err.Error(),
			})
			return
		}

		if req.Name == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Chat name is required",
			})
			return
		}

		chat, err := h.chatService.CreateChat(req.Name, []uuid.UUID{})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to create chat",
				"details": err.Error(),
			})
			return
		}

		location := fmt.Sprintf("/chats/%s", chat.ID.String())
		c.Header("Location", location)

		c.Status(http.StatusCreated)
	}
}
