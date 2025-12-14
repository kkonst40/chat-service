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

func NewChatHandler(newChatService *services.ChatService) *ChatHandler {
	handler := ChatHandler{
		chatService: newChatService,
	}

	return &handler
}

func (h *ChatHandler) GetChat() gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := uuid.Parse(c.Param("id"))
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
		userId, err := uuid.Parse(c.GetString("userID"))
		if err != nil {
			c.Status(http.StatusInternalServerError)
			return
		}

		chats, err := h.chatService.GetChats(userId)
		if err != nil {
			c.Status(http.StatusInternalServerError)
			return
		}
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

		userId, err := uuid.Parse(c.GetString("userID"))
		if err != nil {
			c.Status(http.StatusInternalServerError)
			return
		}

		chat, err := h.chatService.CreateChat(req.Name, []uuid.UUID{userId})
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

func (h *ChatHandler) AddChatUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		//
	}
}

func (h *ChatHandler) UpdateChatName() gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := uuid.Parse(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid chat ID format",
			})
			return
		}

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

		err = h.chatService.UpdateChatName(id, req.Name)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to create chat",
				"details": err.Error(),
			})
			return
		}

		c.Status(http.StatusNoContent)
	}
}

func (h *ChatHandler) DeleteChat() gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := uuid.Parse(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid chat ID format",
			})
			return
		}

		err = h.chatService.DeleteChat(id)
		if err != nil {
			//
		}

		c.Status(http.StatusNoContent)
	}
}
