package handler

import (
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/kkonst40/ichat/internal/api/dto"
	"github.com/kkonst40/ichat/internal/auth"
	errs "github.com/kkonst40/ichat/internal/domain/errors"
	"github.com/kkonst40/ichat/internal/service"
)

type MessageHandler struct {
	messageService *service.MessageService
	validate       *validator.Validate
}

func NewMessageHandler(newMessageService *service.MessageService, validate *validator.Validate) *MessageHandler {
	return &MessageHandler{
		messageService: newMessageService,
		validate:       validate,
	}
}

var defaultFrom = uuid.Nil

const (
	defaultCount int64 = 20
	maxCount     int64 = 100
)

func (h *MessageHandler) GetChatMessages(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	requesterID := auth.GetUserID(ctx)

	var (
		from  uuid.UUID
		count int64
	)

	chatID, err := uuid.Parse(r.PathValue("chatId"))
	if err != nil {
		WriteError(ctx, w, fmt.Errorf("%w: chat ID format", errs.ErrInvalidRequest))
		return
	}

	from, err = uuid.Parse(r.URL.Query().Get("from"))
	if err != nil {
		from = defaultFrom
	}

	count, err = strconv.ParseInt(r.URL.Query().Get("count"), 10, 64)
	if err != nil {
		count = defaultCount
	}

	if count > maxCount {
		count = maxCount
	}

	messages, err := h.messageService.GetChatMessages(ctx, chatID, from, count, requesterID)
	if err != nil {
		WriteError(ctx, w, err)
		return
	}

	slog.DebugContext(ctx, "chat messages retrieved", "chatID", chatID)

	resp := dto.GetMessagesResponse{
		Messages: make([]dto.GetMessageResponse, 0, len(messages)),
	}

	for _, message := range messages {
		resp.Messages = append(resp.Messages, dto.GetMessageResponse{
			ID:        message.ID,
			UserID:    message.UserID,
			UserName:  message.UserName,
			ChatID:    message.ChatID,
			Text:      message.Text,
			CreatedAt: message.CreatedAt,
		})
	}

	WriteJSON(ctx, w, http.StatusOK, resp)
}

func (h *MessageHandler) CreateMessage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	requesterID := auth.GetUserID(ctx)

	chatID, err := uuid.Parse(r.PathValue("chatId"))
	if err != nil {
		WriteError(ctx, w, fmt.Errorf("%w: chat ID format", errs.ErrInvalidRequest))
		return
	}

	var req dto.CreateMessageRequest
	if err := bindJSON(r, &req, h.validate); err != nil {
		WriteError(ctx, w, err)
		return
	}

	if _, err := h.messageService.CreateMessage(ctx, requesterID, chatID, req.Text); err != nil {
		WriteError(ctx, w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *MessageHandler) UpdateMessage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	requesterID := auth.GetUserID(ctx)

	msgID, err := uuid.Parse(r.PathValue("msgId"))
	if err != nil {
		WriteError(ctx, w, fmt.Errorf("%w: message ID format", errs.ErrInvalidRequest))
		return
	}

	var req dto.UpdateMessageRequest
	if err := bindJSON(r, &req, h.validate); err != nil {
		WriteError(ctx, w, err)
		return
	}

	if err := h.messageService.UpdateMessage(ctx, msgID, req.Text, requesterID); err != nil {
		WriteError(ctx, w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *MessageHandler) DeleteMessage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	requesterID := auth.GetUserID(ctx)

	msgID, err := uuid.Parse(r.PathValue("msgId"))
	if err != nil {
		WriteError(ctx, w, fmt.Errorf("%w: message ID format", errs.ErrInvalidRequest))
		return
	}

	if err := h.messageService.DeleteMessage(ctx, msgID, requesterID); err != nil {
		WriteError(ctx, w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
