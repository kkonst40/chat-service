package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/kkonst40/ichat/internal/logger"
)

func DummyAuthQ() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.Query("userId")
		c.Set("requesterID", userID)

		ctx := c.Request.Context()
		logger.FromContext(ctx).Info("", "userID", userID)

		c.Next()
	}
}

func DummyAuthH() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetHeader("UserID")
		c.Set("requesterID", userID)

		ctx := c.Request.Context()
		logger.FromContext(ctx).Info("", "userID", userID)

		c.Next()
	}
}
