package middleware

import (
	"github.com/gin-gonic/gin"
)

func DummyMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetHeader("UserID")
		//if userID == "" {
		//	c.JSON(400, gin.H{"error": "UserID header is required"})
		//	return
		//}

		c.Set("userID", userID)
	}
}
