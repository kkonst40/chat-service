package handlers

import (
	"fmt"
	"log"
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

	log.Println(handler)
	log.Println(&handler)

	return &handler
}

func (h *ChatHandler) GetChat() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		idParam := ctx.Param("id")
		id, err := uuid.Parse(idParam)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid chat ID format",
			})
			return
		}

		chat, err := h.chatService.GetChat(id)
		if err != nil {
			ctx.JSON(http.StatusNotFound, gin.H{
				"error": "Chat not found",
			})
			return
		}

		ctx.JSON(http.StatusOK, chat)
	}
}

func (h *ChatHandler) GetChats() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		chats, _ := h.chatService.GetChats()
		ctx.JSON(http.StatusOK, chats)
	}
}

func (h *ChatHandler) CreateChat() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req struct {
			Name string `json:"name" binding:"required"`
		}

		if err := ctx.ShouldBindJSON(&req); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error":   "Invalid request body",
				"details": err.Error(),
			})
			return
		}

		if req.Name == "" {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error": "Chat name is required",
			})
			return
		}

		chat, err := h.chatService.CreateChat(req.Name, []uuid.UUID{})
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to create chat",
				"details": err.Error(),
			})
			return
		}

		location := fmt.Sprintf("/chats/%s", chat.ID.String())
		ctx.Header("Location", location)

		ctx.Status(http.StatusCreated)
	}
}
