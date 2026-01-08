package handler

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/kkonst40/ichat/internal/apperror"
	"github.com/kkonst40/ichat/internal/dto"
	"github.com/kkonst40/ichat/internal/logger"
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
		ctx := c.Request.Context()

		chatID, err := uuid.Parse(c.Param("chatId"))
		if err != nil {
			c.Error(&apperror.InvalidRequestError{
				Msg: "Invalid chat ID format",
			})
			return
		}

		chat, err := h.chatService.GetChat(ctx, chatID)
		if err != nil {
			c.Error(err)
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
		requesterID := uuid.MustParse(c.GetString("requesterID"))
		ctx := c.Request.Context()

		chats, err := h.chatService.GetUserChats(ctx, requesterID)
		if err != nil {
			c.Error(err)
			return
		}

		logger.FromContext(ctx).Info("user chats retrieved")

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
		requesterID := uuid.MustParse(c.GetString("requesterID"))
		ctx := c.Request.Context()

		if err := c.ShouldBindJSON(&req); err != nil {
			c.Error(handleValidationErr(err))
			return
		}

		chat, err := h.chatService.CreateChat(ctx, req.Name, req.UserIDs, requesterID)
		if err != nil {
			c.Error(err)
			return
		}

		logger.FromContext(ctx).Info("chat created", "chatID", chat.ID)

		location := fmt.Sprintf("/chats/%s", chat.ID.String())
		c.Header("Location", location)

		c.Status(http.StatusCreated)
	}
}

func (h *ChatHandler) UpdateChatName() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req dto.UpdateChatNameRequest
		requesterID := uuid.MustParse(c.GetString("requesterID"))
		ctx := c.Request.Context()

		chatID, err := uuid.Parse(c.Param("chatId"))
		if err != nil {
			c.Error(&apperror.InvalidRequestError{
				Msg: "Invalid chat ID format",
			})
			return
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.Error(handleValidationErr(err))
			return
		}

		err = h.chatService.UpdateChatName(ctx, chatID, req.Name, requesterID)
		if err != nil {
			c.Error(err)
			return
		}

		logger.FromContext(ctx).Info("chat name updated", "chatID", chatID)

		c.Status(http.StatusNoContent)
	}
}

func (h *ChatHandler) DeleteChat() gin.HandlerFunc {
	return func(c *gin.Context) {
		requesterID := uuid.MustParse(c.GetString("requesterID"))
		ctx := c.Request.Context()

		chatID, err := uuid.Parse(c.Param("chatId"))
		if err != nil {
			c.Error(&apperror.InvalidRequestError{
				Msg: "Invalid chat ID format",
			})
			return
		}

		err = h.chatService.DeleteChat(ctx, chatID, requesterID)
		if err != nil {
			c.Error(err)
			return
		}

		logger.FromContext(ctx).Info("chat deleted", "chatID", chatID)

		c.Status(http.StatusNoContent)
	}
}

func (h *ChatHandler) ConnectToChat(wsServer *ws.Server) gin.HandlerFunc {
	return func(c *gin.Context) {
		requesterID := uuid.MustParse(c.GetString("requesterID"))
		ctx := c.Request.Context()

		chatID, err := uuid.Parse(c.Param("chatId"))
		if err != nil {
			c.Error(&apperror.InvalidRequestError{
				Msg: "Invalid chat ID format",
			})
			return
		}

		if !h.chatService.DoesChatExist(ctx, chatID) {
			c.Error(&apperror.NotFoundError{
				Msg: fmt.Sprintf("chat (%v) not found", chatID),
			})
			return
		}

		err = wsServer.Connect(c.Writer, c.Request, requesterID, chatID)
		if err != nil {
			c.Error(err)
			return
		}

		logger.FromContext(ctx).Info("connected to chat", "chatID", chatID)
	}
}

func handleValidationErr(err error) *apperror.InvalidRequestError {
	var ve validator.ValidationErrors
	if errors.As(err, &ve) {
		fields := make([]string, 0, len(ve))
		for _, fe := range ve {
			fields = append(fields, fe.Field())
		}

		return &apperror.InvalidRequestError{
			Msg: "Invalid fields in request body: " + strings.Join(fields, ", "),
		}
	}

	return &apperror.InvalidRequestError{
		Msg: "Invalid request body",
	}
}
