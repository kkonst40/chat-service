package handler

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/kkonst40/ichat/internal/dto"
	errs "github.com/kkonst40/ichat/internal/errors"
	"github.com/kkonst40/ichat/internal/logger"
	"github.com/kkonst40/ichat/internal/service"
)

type MessageHandler struct {
	messageService *service.MessageService
}

func NewMessageHandler(newMessageService *service.MessageService) *MessageHandler {
	return &MessageHandler{
		messageService: newMessageService,
	}
}

func (h *MessageHandler) GetChatMessages(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	requesterID := getUserID(ctx)
	log := logger.FromContext(ctx)

	chatID, err := uuid.Parse(r.PathValue("chatId"))
	if err != nil {
		WriteError(w, fmt.Errorf("%w: chat ID format", errs.ErrInvalidRequest), log)
		return
	}

	from, err := strconv.ParseInt(r.PathValue("from"), 10, 64)
	if err != nil {
		WriteError(w, fmt.Errorf("%w: 'from' param", errs.ErrInvalidRequest), log)
		return
	}

	count, err := strconv.ParseInt(r.PathValue("count"), 10, 64)
	if err != nil {
		WriteError(w, fmt.Errorf("%w: 'count' param", errs.ErrInvalidRequest), log)
		return
	}

	messages, err := h.messageService.GetChatMessages(ctx, chatID, from, count, requesterID)
	if err != nil {
		WriteError(w, err, log)
		return
	}

	log.Info("chat messages retrieved", "chatID", chatID)

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

	WriteJSON(w, http.StatusOK, resp, log)
}
