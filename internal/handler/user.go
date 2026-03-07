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

type UserHandler struct {
	userService *service.UserService
	validate    *validator.Validate
}

func NewUserHandler(newUserService *service.UserService, validate *validator.Validate) *UserHandler {
	return &UserHandler{
		userService: newUserService,
		validate:    validate,
	}
}

func (h *UserHandler) GetChatUsers(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	requesterID := auth.GetUserID(ctx)

	chatID, err := uuid.Parse(r.PathValue("chatId"))
	if err != nil {
		WriteError(ctx, w, fmt.Errorf("%w: chat ID format", errs.ErrInvalidRequest))
		return
	}

	users, err := h.userService.GetChatUsers(ctx, chatID, requesterID)
	if err != nil {
		WriteError(ctx, w, err)
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

	WriteJSON(ctx, w, http.StatusOK, resp)
}

func (h *UserHandler) AddChatUsers(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	requesterID := auth.GetUserID(ctx)

	chatID, err := uuid.Parse(r.PathValue("chatId"))
	if err != nil {
		WriteError(ctx, w, fmt.Errorf("%w: chat ID format", errs.ErrInvalidRequest))
		return
	}

	var req dto.AddChatUsersRequest
	if err := bindJSON(r, &req, h.validate); err != nil {
		WriteError(ctx, w, err)
		return
	}

	err = h.userService.AddChatUsers(ctx, chatID, req.UserNames, requesterID)
	if err != nil {
		WriteError(ctx, w, err)
		return
	}

	slog.DebugContext(ctx, "chat users added", "chatID", chatID)
}

func (h *UserHandler) UpdateChatUserRole(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	requesterID := auth.GetUserID(ctx)

	chatID, err := uuid.Parse(r.PathValue("chatId"))
	if err != nil {
		WriteError(ctx, w, fmt.Errorf("%w: chat ID format", errs.ErrInvalidRequest))
		return
	}

	userID, err := uuid.Parse(r.PathValue("userId"))
	if err != nil {
		WriteError(ctx, w, fmt.Errorf("%w: user ID format", errs.ErrInvalidRequest))
		return
	}

	var req dto.UpdateChatUserRoleRequest
	if err := bindJSON(r, &req, h.validate); err != nil {
		WriteError(ctx, w, err)
		return
	}

	switch req.Role {
	case model.Common, model.Admin, model.Owner:
	default:
		WriteError(ctx, w, fmt.Errorf("%w: role name", errs.ErrInvalidRequest))
		return
	}

	err = h.userService.UpdateUserRole(ctx, chatID, userID, req.Role, requesterID)
	if err != nil {
		WriteError(ctx, w, err)
		return
	}

	slog.DebugContext(ctx, "chat user role updated", "chatID", chatID, "userID", userID)
}

func (h *UserHandler) DeleteChatUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	requesterID := auth.GetUserID(ctx)

	chatID, err := uuid.Parse(r.PathValue("chatId"))
	if err != nil {
		WriteError(ctx, w, fmt.Errorf("%w: chat ID format", errs.ErrInvalidRequest))
		return
	}

	userID, err := uuid.Parse(r.PathValue("userId"))
	if err != nil {
		WriteError(ctx, w, fmt.Errorf("%w: user ID format", errs.ErrInvalidRequest))
		return
	}

	err = h.userService.DeleteChatUser(ctx, chatID, userID, requesterID)
	if err != nil {
		WriteError(ctx, w, err)
		return
	}

	slog.DebugContext(ctx, "chat user deleted", "chatID", chatID, "userID", userID)
}
