package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/kkonst40/ichat/internal/apperror"
)

func ErrorMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		_ = apperror.ErrChatAlreadyExists
		c.Next()
	}
}
