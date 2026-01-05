package middleware

import (
	"context"
	"log/slog"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/kkonst40/ichat/internal/logger"
)

func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID, _ := uuid.NewV7()
		log := slog.With("requestID", requestID.String())

		log.Info("Request started",
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
		)

		ctx := context.WithValue(c.Request.Context(), logger.CtxKey, log)

		c.Request = c.Request.WithContext(ctx)
		c.Next()

		c.Writer.Status()
	}
}
