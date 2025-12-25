package handler

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/kkonst40/ichat/internal/dto"
	"github.com/kkonst40/ichat/internal/service"
	"github.com/kkonst40/ichat/internal/ws"
)

type ChatHandler struct {
	chatService *service.ChatService
}

func NewChatHandler(newChatService *service.ChatService) *ChatHandler {
	handler := ChatHandler{
		chatService: newChatService,
	}

	return &handler
}

func (h *ChatHandler) GetChat() gin.HandlerFunc {
	return func(c *gin.Context) {
		ID, err := uuid.Parse(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid chat ID format",
			})
			return
		}

		chat, err := h.chatService.GetChat(ID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Chat not found",
			})
			return
		}

		resp := dto.GetChatResponse{
			ID:   chat.ID,
			Name: chat.Name,
		}

		c.JSON(http.StatusOK, resp)
	}
}

func (h *ChatHandler) GetChats() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, err := uuid.Parse(c.GetString("userID"))
		if err != nil {
			c.Status(http.StatusInternalServerError)
			return
		}

		chats, err := h.chatService.GetChats(userID)
		if err != nil {
			//
			return
		}

		resp := dto.GetChatsResponse{
			Chats: make([]dto.GetChatResponse, 0, len(chats)),
		}

		for _, chat := range chats {
			resp.Chats = append(resp.Chats, dto.GetChatResponse{
				ID:   chat.ID,
				Name: chat.Name,
			})
		}

		c.JSON(http.StatusOK, resp)
	}
}

func (h *ChatHandler) CreateChat() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req dto.CreateChatRequest

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

		userID, err := uuid.Parse(c.GetString("userID"))
		if err != nil {
			c.Status(http.StatusInternalServerError)
			return
		}

		chat, err := h.chatService.CreateChat(req.Name, req.UserIDs, userID)
		if err != nil {
			//
			return
		}

		location := fmt.Sprintf("/chats/%s", chat.ID.String())
		c.Header("Location", location)

		c.Status(http.StatusCreated)
	}
}

func (h *ChatHandler) UpdateChatName() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req dto.UpdateChatNameRequest

		chatID, err := uuid.Parse(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid chat ID format",
			})
			return
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

		userID := uuid.MustParse(c.GetString("userID"))

		err = h.chatService.UpdateChatName(chatID, req.Name, userID)
		if err != nil {
			//
			return
		}

		c.Status(http.StatusNoContent)
	}
}

func (h *ChatHandler) DeleteChat() gin.HandlerFunc {
	return func(c *gin.Context) {
		chatID, err := uuid.Parse(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid chat ID format",
			})
			return
		}

		userID := uuid.MustParse(c.GetString("userID"))

		err = h.chatService.DeleteChat(chatID, userID)
		if err != nil {
			//
			return
		}

		c.Status(http.StatusNoContent)
	}
}

func (h *ChatHandler) ConnectToChat(wsServer *ws.Server) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, err := uuid.Parse(c.GetString("userID"))
		if err != nil {
			c.Status(http.StatusInternalServerError)
			return
		}

		chatID, err := uuid.Parse(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid chat ID format",
			})
			return
		}

		err = wsServer.Connect(c.Writer, c.Request, userID, chatID)
		if err != nil {
			//
			return
		}
	}
}
