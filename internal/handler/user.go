package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/kkonst40/ichat/internal/apperror"
	"github.com/kkonst40/ichat/internal/dto"
	"github.com/kkonst40/ichat/internal/logger"
	"github.com/kkonst40/ichat/internal/model"
	"github.com/kkonst40/ichat/internal/service"
)

type UserHandler struct {
	userService *service.UserService
}

func NewUserHandler(newUserService *service.UserService) *UserHandler {
	return &UserHandler{
		userService: newUserService,
	}
}

func (h *UserHandler) GetChatUsers() gin.HandlerFunc {
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

		users, err := h.userService.GetChatUsers(ctx, chatID, requesterID)
		if err != nil {
			c.Error(err)
			return
		}

		logger.FromContext(ctx).Info("chat users retrieved", "chatID", chatID)

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

		c.JSON(http.StatusOK, resp)
	}
}

func (h *UserHandler) AddChatUsers() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req dto.AddChatUsersRequest
		requesterID := uuid.MustParse(c.GetString("requesterID"))
		ctx := c.Request.Context()

		chatID, err := uuid.Parse(c.Param("chatId"))
		if err != nil {
			c.Error(&apperror.InvalidRequestError{
				Msg: "Invalid chat ID format",
			})
			return
		}

		if err = c.ShouldBindJSON(&req); err != nil {
			c.Error(validationErr(err))
			return
		}

		err = h.userService.AddChatUsers(ctx, chatID, req.UserIDs, requesterID)
		if err != nil {
			c.Error(err)
			return
		}

		logger.FromContext(ctx).Info("chat users added", "chatID", chatID)
	}
}

func (h *UserHandler) UpdateChatUserRole() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req dto.UpdateChatUserRoleRequest
		requesterID := uuid.MustParse(c.GetString("requesterID"))
		ctx := c.Request.Context()

		chatID, err := uuid.Parse(c.Param("chatId"))
		if err != nil {
			c.Error(&apperror.InvalidRequestError{
				Msg: "Invalid chat ID format",
			})
			return
		}

		userID, err := uuid.Parse(c.Param("userId"))
		if err != nil {
			c.Error(&apperror.InvalidRequestError{
				Msg: "Invalid user ID format",
			})
			return
		}

		if err = c.ShouldBindJSON(&req); err != nil {
			c.Error(validationErr(err))
			return
		}

		var role model.Role
		switch req.Role {
		case "common":
			role = model.Common
		case "admin":
			role = model.Admin
		case "owner":
			role = model.Owner
		default:
			c.Error(&apperror.InvalidRequestError{
				Msg: "Invalid user role name",
			})
			return
		}

		err = h.userService.UpdateUserRole(ctx, chatID, userID, role, requesterID)
		if err != nil {
			c.Error(err)
			return
		}

		logger.FromContext(ctx).Info("chat user role updated", "chatID", chatID, "userID", userID)
	}
}

func (h *UserHandler) DeleteChatUser() gin.HandlerFunc {
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

		userID, err := uuid.Parse(c.Param("userId"))
		if err != nil {
			c.Error(&apperror.InvalidRequestError{
				Msg: "Invalid user ID format",
			})
			return
		}

		err = h.userService.DeleteChatUser(ctx, chatID, userID, requesterID)
		if err != nil {
			c.Error(err)
			return
		}

		logger.FromContext(ctx).Info("chat user deleted", "chatID", chatID, "userID", userID)
	}
}
