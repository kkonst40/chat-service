package handler

import (
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/kkonst40/ichat/internal/dto"
	errs "github.com/kkonst40/ichat/internal/errors"
	"github.com/kkonst40/ichat/internal/logger"
	"github.com/kkonst40/ichat/internal/model"
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
	requesterID := getUserID(ctx)
	log := logger.FromContext(ctx)

	chatID, err := uuid.Parse(r.PathValue("chatId"))
	if err != nil {
		WriteString(w, http.StatusBadRequest, "Invalid chat ID format", log)
		return
	}

	users, err := h.userService.GetChatUsers(ctx, chatID, requesterID)
	if err != nil {
		statusCode, resp := errs.MapError(err, log)
		WriteJSON(w, statusCode, resp, log)
		return
	}

	log.Info("chat users retrieved", "chatID", chatID)

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

	WriteJSON(w, http.StatusOK, resp, log)
}

func (h *UserHandler) AddChatUsers(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	requesterID := getUserID(ctx)
	log := logger.FromContext(ctx)

	chatID, err := uuid.Parse(r.PathValue("chatId"))
	if err != nil {
		WriteString(w, http.StatusBadRequest, "Invalid chat ID format", log)
		return
	}

	var req dto.AddChatUsersRequest
	if err := bindJSON(r, &req, h.validate); err != nil {
		statusCode, resp := errs.MapError(handleValidationErr(err), log)
		WriteJSON(w, statusCode, resp, log)
		return
	}

	err = h.userService.AddChatUsers(ctx, chatID, req.UserIDs, requesterID)
	if err != nil {
		statusCode, resp := errs.MapError(err, log)
		WriteJSON(w, statusCode, resp, log)
		return
	}

	log.Info("chat users added", "chatID", chatID)
}

func (h *UserHandler) UpdateChatUserRole(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	requesterID := getUserID(ctx)
	log := logger.FromContext(ctx)

	chatID, err := uuid.Parse(r.PathValue("chatId"))
	if err != nil {
		WriteString(w, http.StatusBadRequest, "Invalid chat ID format", log)
		return
	}

	userID, err := uuid.Parse(r.PathValue("userId"))
	if err != nil {
		WriteString(w, http.StatusBadRequest, "Invalid user ID format", log)
		return
	}

	var req dto.UpdateChatUserRoleRequest
	if err := bindJSON(r, &req, h.validate); err != nil {
		statusCode, resp := errs.MapError(handleValidationErr(err), log)
		WriteJSON(w, statusCode, resp, log)
		return
	}

	switch req.Role {
	case model.Common, model.Admin, model.Owner:
	default:
		WriteString(w, http.StatusBadRequest, "Invalid user role name", log)
		return
	}

	err = h.userService.UpdateUserRole(ctx, chatID, userID, req.Role, requesterID)
	if err != nil {
		statusCode, resp := errs.MapError(err, log)
		WriteJSON(w, statusCode, resp, log)
		return
	}

	log.Info("chat user role updated", "chatID", chatID, "userID", userID)
}

func (h *UserHandler) DeleteChatUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	requesterID := getUserID(ctx)
	log := logger.FromContext(ctx)

	chatID, err := uuid.Parse(r.PathValue("chatId"))
	if err != nil {
		WriteString(w, http.StatusBadRequest, "Invalid chat ID format", log)
		return
	}

	userID, err := uuid.Parse(r.PathValue("userId"))
	if err != nil {
		WriteString(w, http.StatusBadRequest, "Invalid user ID format", log)
		return
	}

	err = h.userService.DeleteChatUser(ctx, chatID, userID, requesterID)
	if err != nil {
		statusCode, resp := errs.MapError(err, log)
		WriteJSON(w, statusCode, resp, log)
		return
	}

	log.Info("chat user deleted", "chatID", chatID, "userID", userID)
}
