package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/kkonst40/ichat/internal/dto"
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
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid chat ID format",
			})
			return
		}

		users, err := h.userService.GetChatUsers(ctx, chatID, requesterID)
		if err != nil {
			//
			return
		}

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
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid chat ID format",
			})
			return
		}

		err = c.ShouldBindJSON(&req)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Invalid request body",
				"details": err.Error(),
			})
			return
		}

		err = h.userService.AddChatUsers(ctx, chatID, req.UserIDs, requesterID)
		if err != nil {
			//
			return
		}
	}
}

func (h *UserHandler) SetChatUserRole() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req dto.UpdateChatUserRoleRequest
		requesterID := uuid.MustParse(c.GetString("requesterID"))
		ctx := c.Request.Context()

		chatID, err := uuid.Parse(c.Param("chatId"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid chat ID format",
			})
			return
		}

		userID, err := uuid.Parse(c.Param("userId"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid chat ID format",
			})
			return
		}

		err = c.ShouldBindJSON(&req)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Invalid request body",
				"details": err.Error(),
			})
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
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid role name",
			})
			return
		}

		err = h.userService.SetUserRole(ctx, chatID, userID, role, requesterID)
		if err != nil {
			//
			return
		}
	}
}

func (h *UserHandler) DeleteChatUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		requesterID := uuid.MustParse(c.GetString("requesterID"))
		ctx := c.Request.Context()

		chatID, err := uuid.Parse(c.Param("chatId"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid chat ID format",
			})
			return
		}

		userID, err := uuid.Parse(c.Param("userId"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid chat ID format",
			})
			return
		}

		err = h.userService.DeleteChatUser(ctx, chatID, userID, requesterID)
		if err != nil {
			//
			return
		}
	}
}
