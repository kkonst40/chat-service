package middleware

import (
	"github.com/gin-gonic/gin"
)

func DummyAuthQ() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.Query("userId")
		c.Set("requesterID", userID)
		c.Next()
	}
}

func DummyAuthH() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetHeader("UserID")
		c.Set("requesterID", userID)
		c.Next()
	}
}
