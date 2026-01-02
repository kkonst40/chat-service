package middleware

import (
	"log"

	"github.com/gin-gonic/gin"
)

func Dummy() gin.HandlerFunc {
	return func(c *gin.Context) {
		//userID := c.GetHeader("UserID")
		userID := c.Query("userId")
		//if userID == "" {
		//	c.JSON(400, gin.H{"error": "UserID header is required"})
		//	return
		//}
		log.Println(userID)
		c.Set("userID", userID)
		c.Next()
	}
}
