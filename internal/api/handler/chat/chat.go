package chat

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/kkonst40/chat-service/internal/api/dto"
	"github.com/kkonst40/chat-service/internal/api/handler"
	errs "github.com/kkonst40/chat-service/internal/domain/errors"
	"github.com/kkonst40/chat-service/internal/domain/model"
	"github.com/kkonst40/chat-service/internal/service/auth"
)

type Handler struct {
	chatService ChatService
	validate    *validator.Validate
}

type ChatService interface {
	GetChat(ctx context.Context, chatID uuid.UUID) (model.Chat, error)
	GetUserChats(ctx context.Context, userID uuid.UUID, filter model.ChatFilter) ([]model.Chat, error)
	CreatePersonalChat(ctx context.Context, userID1 uuid.UUID, userName2 string) (model.Chat, error)
	CreateGroupChat(ctx context.Context, name string, userNames []string, requesterID uuid.UUID) (model.Chat, error)
	UpdateChatName(ctx context.Context, chatID uuid.UUID, name string, requesterID uuid.UUID) error
	DeleteChat(ctx context.Context, chatID uuid.UUID, requesterID uuid.UUID) error
}

func New(newChatService ChatService, validate *validator.Validate) *Handler {
	return &Handler{
		chatService: newChatService,
		validate:    validate,
	}
}

func (h *Handler) GetChat(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	chatID, err := uuid.Parse(r.PathValue("chatId"))
	if err != nil {
		handler.WriteError(ctx, w, fmt.Errorf("%w: chat ID format", errs.ErrInvalidRequest))
		return
	}

	chat, err := h.chatService.GetChat(ctx, chatID)
	if err != nil {
		handler.WriteError(ctx, w, err)
		return
	}

	resp := dto.GetChatResponse{
		ID:            chat.ID,
		Name:          chat.Name,
		IsGroup:       chat.IsGroup,
		LastMessageAt: chat.LastMessageAt,
	}

	handler.WriteJSON(ctx, w, http.StatusOK, resp)
}

func (h *Handler) GetChats(w http.ResponseWriter, r *http.Request) {
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
		handler.WriteError(ctx, w, err)
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

	handler.WriteJSON(ctx, w, http.StatusOK, resp)
}

func (h *Handler) CreateGroupChat(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	requesterID := auth.GetUserID(ctx)

	var req dto.CreateGroupChatRequest
	if err := handler.BindJSON(r, &req, h.validate); err != nil {
		handler.WriteError(ctx, w, err)
		return
	}

	chat, err := h.chatService.CreateGroupChat(ctx, req.Name, req.UserNames, requesterID)
	if err != nil {
		handler.WriteError(ctx, w, err)
		return
	}

	slog.DebugContext(ctx, "chat created", "chatID", chat.ID)
	location := fmt.Sprintf("/chats/%s", chat.ID.String())

	w.Header().Set("Location", location)
	w.WriteHeader(http.StatusCreated)
}

func (h *Handler) CreatePersonalChat(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	requesterID := auth.GetUserID(ctx)

	var req dto.CreatePersonalChatRequest
	if err := handler.BindJSON(r, &req, h.validate); err != nil {
		handler.WriteError(ctx, w, err)
		return
	}

	chat, err := h.chatService.CreatePersonalChat(ctx, requesterID, req.UserName)
	if err != nil {
		handler.WriteError(ctx, w, err)
		return
	}

	slog.DebugContext(ctx, "chat created", "chatID", chat.ID)
	location := fmt.Sprintf("/chats/%s", chat.ID.String())

	w.Header().Set("Location", location)
	w.WriteHeader(http.StatusCreated)
}

func (h *Handler) UpdateChatName(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	requesterID := auth.GetUserID(ctx)

	chatID, err := uuid.Parse(r.PathValue("chatId"))
	if err != nil {
		handler.WriteError(ctx, w, fmt.Errorf("%w: chat ID format", errs.ErrInvalidRequest))
		return
	}

	var req dto.UpdateChatNameRequest
	if err := handler.BindJSON(r, &req, h.validate); err != nil {
		handler.WriteError(ctx, w, handler.HandleValidationErr(err))
		return
	}

	err = h.chatService.UpdateChatName(ctx, chatID, req.Name, requesterID)
	if err != nil {
		handler.WriteError(ctx, w, err)
		return
	}

	slog.DebugContext(ctx, "chat name updated", "chatID", chatID)
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) DeleteChat(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	requesterID := auth.GetUserID(ctx)

	chatID, err := uuid.Parse(r.PathValue("chatId"))
	if err != nil {
		handler.WriteError(ctx, w, fmt.Errorf("%w: chat ID format", errs.ErrInvalidRequest))
		return
	}

	err = h.chatService.DeleteChat(ctx, chatID, requesterID)
	if err != nil {
		handler.WriteError(ctx, w, err)
		return
	}

	slog.DebugContext(ctx, "chat deleted", "chatID", chatID)

	w.WriteHeader(http.StatusNoContent)
}
