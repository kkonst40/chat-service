package handler

import (
	"fmt"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	errs "github.com/kkonst40/ichat/internal/domain/errors"
	"github.com/kkonst40/ichat/internal/domain/model"
	"github.com/kkonst40/ichat/internal/dto"
	"github.com/kkonst40/ichat/internal/logger"
	"github.com/kkonst40/ichat/internal/service"
)

type ChatHandler struct {
	chatService *service.ChatService
	validate    *validator.Validate
}

func NewChatHandler(newChatService *service.ChatService, validate *validator.Validate) *ChatHandler {
	return &ChatHandler{
		chatService: newChatService,
		validate:    validate,
	}
}

func (h *ChatHandler) GetChat(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.FromContext(ctx)

	chatID, err := uuid.Parse(r.PathValue("chatId"))
	if err != nil {
		WriteError(w, fmt.Errorf("%w: chat ID format", errs.ErrInvalidRequest), log)
		return
	}

	chat, err := h.chatService.GetChat(ctx, chatID)
	if err != nil {
		WriteError(w, err, log)
		return
	}

	resp := dto.GetChatResponse{
		ID:   chat.ID,
		Name: chat.Name,
	}

	WriteJSON(w, http.StatusOK, resp, log)
}

func (h *ChatHandler) GetChats(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	requesterID := getUserID(ctx)
	log := logger.FromContext(ctx)

	var chats []model.Chat
	var err error

	switch r.URL.Query().Get("filter") {
	case "":
		chats, err = h.chatService.GetUserChats(ctx, requesterID, model.AllChats)
	case "personal":
		chats, err = h.chatService.GetUserChats(ctx, requesterID, model.PersonalChats)
	case "group":
		chats, err = h.chatService.GetUserChats(ctx, requesterID, model.GroupChats)
	}

	if err != nil {
		WriteError(w, err, log)
		return
	}

	log.Debug("user chats retrieved")

	resp := dto.GetChatsResponse{
		Chats: make([]dto.GetChatResponse, 0, len(chats)),
	}

	for _, chat := range chats {
		resp.Chats = append(resp.Chats, dto.GetChatResponse{
			ID:            chat.ID,
			Name:          chat.Name,
			LastMessageAt: chat.LastMessageAt,
		})
	}

	WriteJSON(w, http.StatusOK, resp, log)
}

func (h *ChatHandler) CreateChat(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	requesterID := getUserID(ctx)
	log := logger.FromContext(ctx)

	var req dto.CreateChatRequest
	if err := bindJSON(r, &req, h.validate); err != nil {
		WriteError(w, err, log)
		return
	}

	chat, err := h.chatService.CreateChat(ctx, req.Name, req.UserIDs, requesterID)
	if err != nil {
		WriteError(w, err, log)
		return
	}

	log.Debug("chat created", "chatID", chat.ID)
	location := fmt.Sprintf("/chats/%s", chat.ID.String())

	w.Header().Set("Location", location)
	w.WriteHeader(http.StatusCreated)
}

func (h *ChatHandler) UpdateChatName(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	requesterID := getUserID(ctx)
	log := logger.FromContext(ctx)

	chatID, err := uuid.Parse(r.PathValue("chatId"))
	if err != nil {
		WriteError(w, fmt.Errorf("%w: chat ID format", errs.ErrInvalidRequest), log)
		return
	}

	var req dto.UpdateChatNameRequest
	if err := bindJSON(r, &req, h.validate); err != nil {
		WriteError(w, handleValidationErr(err), log)
		return
	}

	err = h.chatService.UpdateChatName(ctx, chatID, req.Name, requesterID)
	if err != nil {
		WriteError(w, err, log)
		return
	}

	log.Debug("chat name updated", "chatID", chatID)
	w.WriteHeader(http.StatusNoContent)
}

func (h *ChatHandler) DeleteChat(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	requesterID := getUserID(ctx)
	log := logger.FromContext(ctx)

	chatID, err := uuid.Parse(r.PathValue("chatId"))
	if err != nil {
		WriteError(w, fmt.Errorf("%w: chat ID format", errs.ErrInvalidRequest), log)
		return
	}

	err = h.chatService.DeleteChat(ctx, chatID, requesterID)
	if err != nil {
		WriteError(w, err, log)
		return
	}

	log.Debug("chat deleted", "chatID", chatID)

	w.WriteHeader(http.StatusNoContent)
}
