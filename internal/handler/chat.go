package handler

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/kkonst40/ichat/internal/auth"
	errs "github.com/kkonst40/ichat/internal/domain/errors"
	"github.com/kkonst40/ichat/internal/domain/model"
	"github.com/kkonst40/ichat/internal/dto"
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

	chatID, err := uuid.Parse(r.PathValue("chatId"))
	if err != nil {
		WriteError(ctx, w, fmt.Errorf("%w: chat ID format", errs.ErrInvalidRequest))
		return
	}

	chat, err := h.chatService.GetChat(ctx, chatID)
	if err != nil {
		WriteError(ctx, w, err)
		return
	}

	resp := dto.GetChatResponse{
		ID:            chat.ID,
		Name:          chat.Name,
		IsGroup:       chat.IsGroup,
		LastMessageAt: chat.LastMessageAt,
	}

	WriteJSON(ctx, w, http.StatusOK, resp)
}

func (h *ChatHandler) GetChats(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	requesterID := auth.GetUserID(ctx)

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
		WriteError(ctx, w, err)
		return
	}

	slog.DebugContext(ctx, "user chats retrieved")

	resp := dto.GetChatsResponse{
		Chats: make([]dto.GetChatResponse, 0, len(chats)),
	}

	for _, chat := range chats {
		resp.Chats = append(resp.Chats, dto.GetChatResponse{
			ID:            chat.ID,
			Name:          chat.Name,
			IsGroup:       chat.IsGroup,
			LastMessageAt: chat.LastMessageAt,
		})
	}

	WriteJSON(ctx, w, http.StatusOK, resp)
}

func (h *ChatHandler) CreateGroupChat(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	requesterID := auth.GetUserID(ctx)

	var req dto.CreateGroupChatRequest
	if err := bindJSON(r, &req, h.validate); err != nil {
		WriteError(ctx, w, err)
		return
	}

	chat, err := h.chatService.CreateGroupChat(ctx, req.Name, req.UserNames, requesterID)
	if err != nil {
		WriteError(ctx, w, err)
		return
	}

	slog.DebugContext(ctx, "chat created", "chatID", chat.ID)
	location := fmt.Sprintf("/chats/%s", chat.ID.String())

	w.Header().Set("Location", location)
	w.WriteHeader(http.StatusCreated)
}

func (h *ChatHandler) CreatePersonalChat(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	requesterID := auth.GetUserID(ctx)

	var req dto.CreatePersonalChatRequest
	if err := bindJSON(r, &req, h.validate); err != nil {
		WriteError(ctx, w, err)
		return
	}

	chat, err := h.chatService.CreatePersonalChat(ctx, requesterID, req.UserName)
	if err != nil {
		WriteError(ctx, w, err)
		return
	}

	slog.DebugContext(ctx, "chat created", "chatID", chat.ID)
	location := fmt.Sprintf("/chats/%s", chat.ID.String())

	w.Header().Set("Location", location)
	w.WriteHeader(http.StatusCreated)
}

func (h *ChatHandler) UpdateChatName(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	requesterID := auth.GetUserID(ctx)

	chatID, err := uuid.Parse(r.PathValue("chatId"))
	if err != nil {
		WriteError(ctx, w, fmt.Errorf("%w: chat ID format", errs.ErrInvalidRequest))
		return
	}

	var req dto.UpdateChatNameRequest
	if err := bindJSON(r, &req, h.validate); err != nil {
		WriteError(ctx, w, handleValidationErr(err))
		return
	}

	err = h.chatService.UpdateChatName(ctx, chatID, req.Name, requesterID)
	if err != nil {
		WriteError(ctx, w, err)
		return
	}

	slog.DebugContext(ctx, "chat name updated", "chatID", chatID)
	w.WriteHeader(http.StatusNoContent)
}

func (h *ChatHandler) DeleteChat(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	requesterID := auth.GetUserID(ctx)

	chatID, err := uuid.Parse(r.PathValue("chatId"))
	if err != nil {
		WriteError(ctx, w, fmt.Errorf("%w: chat ID format", errs.ErrInvalidRequest))
		return
	}

	err = h.chatService.DeleteChat(ctx, chatID, requesterID)
	if err != nil {
		WriteError(ctx, w, err)
		return
	}

	slog.DebugContext(ctx, "chat deleted", "chatID", chatID)

	w.WriteHeader(http.StatusNoContent)
}
