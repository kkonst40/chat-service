package user

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
	userService UserService
	validate    *validator.Validate
}

type UserService interface {
	GetChatUsers(ctx context.Context, chatID uuid.UUID, requesterID uuid.UUID) ([]model.User, error)
	AddChatUsers(ctx context.Context, chatID uuid.UUID, userNames []string, requesterID uuid.UUID) error
	UpdateUserRole(ctx context.Context, chatID uuid.UUID, userID uuid.UUID, newRole model.Role, requesterID uuid.UUID) error
	DeleteChatUser(ctx context.Context, chatID uuid.UUID, userID uuid.UUID, requesterID uuid.UUID) error
}

func New(newUserService UserService, validate *validator.Validate) *Handler {
	return &Handler{
		userService: newUserService,
		validate:    validate,
	}
}

func (h *Handler) GetChatUsers(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	requesterID := auth.GetUserID(ctx)

	chatID, err := uuid.Parse(r.PathValue("chatId"))
	if err != nil {
		handler.WriteError(ctx, w, fmt.Errorf("%w: chat ID format", errs.ErrInvalidRequest))
		return
	}

	users, err := h.userService.GetChatUsers(ctx, chatID, requesterID)
	if err != nil {
		handler.WriteError(ctx, w, err)
		return
	}

	slog.DebugContext(ctx, "chat users retrieved", "chatID", chatID)

	resp := dto.GetChatUsersResponse{
		Users: make([]dto.GetUserResponse, 0, len(users)),
	}

	for _, user := range users {
		resp.Users = append(resp.Users, dto.GetUserResponse{
			ID:     user.ID,
			ChatID: user.ChatID,
			Role:   string(user.Role),
		})
	}

	handler.WriteJSON(ctx, w, http.StatusOK, resp)
}

func (h *Handler) AddChatUsers(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	requesterID := auth.GetUserID(ctx)

	chatID, err := uuid.Parse(r.PathValue("chatId"))
	if err != nil {
		handler.WriteError(ctx, w, fmt.Errorf("%w: chat ID format", errs.ErrInvalidRequest))
		return
	}

	var req dto.AddChatUsersRequest
	if err := handler.BindJSON(r, &req, h.validate); err != nil {
		handler.WriteError(ctx, w, err)
		return
	}

	err = h.userService.AddChatUsers(ctx, chatID, req.UserNames, requesterID)
	if err != nil {
		handler.WriteError(ctx, w, err)
		return
	}

	slog.DebugContext(ctx, "chat users added", "chatID", chatID)
}

func (h *Handler) UpdateChatUserRole(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	requesterID := auth.GetUserID(ctx)

	chatID, err := uuid.Parse(r.PathValue("chatId"))
	if err != nil {
		handler.WriteError(ctx, w, fmt.Errorf("%w: chat ID format", errs.ErrInvalidRequest))
		return
	}

	userID, err := uuid.Parse(r.PathValue("userId"))
	if err != nil {
		handler.WriteError(ctx, w, fmt.Errorf("%w: user ID format", errs.ErrInvalidRequest))
		return
	}

	var req dto.UpdateChatUserRoleRequest
	if err := handler.BindJSON(r, &req, h.validate); err != nil {
		handler.WriteError(ctx, w, err)
		return
	}

	switch req.Role {
	case model.Common, model.Admin, model.Owner:
	default:
		handler.WriteError(ctx, w, fmt.Errorf("%w: role name", errs.ErrInvalidRequest))
		return
	}

	err = h.userService.UpdateUserRole(ctx, chatID, userID, req.Role, requesterID)
	if err != nil {
		handler.WriteError(ctx, w, err)
		return
	}

	slog.DebugContext(ctx, "chat user role updated", "chatID", chatID, "userID", userID)
}

func (h *Handler) DeleteChatUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	requesterID := auth.GetUserID(ctx)

	chatID, err := uuid.Parse(r.PathValue("chatId"))
	if err != nil {
		handler.WriteError(ctx, w, fmt.Errorf("%w: chat ID format", errs.ErrInvalidRequest))
		return
	}

	userID, err := uuid.Parse(r.PathValue("userId"))
	if err != nil {
		handler.WriteError(ctx, w, fmt.Errorf("%w: user ID format", errs.ErrInvalidRequest))
		return
	}

	err = h.userService.DeleteChatUser(ctx, chatID, userID, requesterID)
	if err != nil {
		handler.WriteError(ctx, w, err)
		return
	}

	slog.DebugContext(ctx, "chat user deleted", "chatID", chatID, "userID", userID)
}
