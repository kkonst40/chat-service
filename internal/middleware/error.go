package middleware

import (
	"context"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kkonst40/ichat/internal/apperror"
	"github.com/kkonst40/ichat/internal/logger"
)

func Error() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if len(c.Errors) > 0 {
			err := c.Errors.Last()

			var statusCode int
			var message string
			log := logger.FromContext(c.Request.Context())

			var (
				nfErr       *apperror.NotFoundError
				invReqErr   *apperror.InvalidRequestError
				frbErr      *apperror.ForbiddenError
				unauthErr   *apperror.UnauthorizedError
				dbErr       *apperror.DBError
				chatConnErr *apperror.ChatConnectionError
			)

			switch {
			case errors.Is(err, context.DeadlineExceeded):
				statusCode = http.StatusGatewayTimeout
				message = "The request took too long to process"
				log.Error("Context deadline exceeded", "errors", err)

			case errors.As(err, &nfErr):
				statusCode = http.StatusNotFound
				message = nfErr.Msg
				log.Error("Resource not found", "errors", err.Error())

			case errors.As(err, &invReqErr):
				statusCode = http.StatusBadRequest
				message = invReqErr.Msg
				log.Error("Invalid request", "errors", err.Error())

			case errors.As(err, &frbErr):
				statusCode = http.StatusForbidden
				message = "Access denied"
				log.Error("User has no access", "errors", err.Error())

			case errors.As(err, &unauthErr):
				statusCode = http.StatusUnauthorized
				message = "User unauthorized"
				log.Error("User authorization failed", "errors", err.Error())

			case errors.As(err, &dbErr):
				statusCode = http.StatusInternalServerError
				message = "Internal server error"
				log.Error("Database connection error while executing query", "errors", err.Error())

			case errors.As(err, &chatConnErr):
				statusCode = http.StatusForbidden
				message = "Upgrading to WebSocket error"
				log.Error("Upgrading from HTTP to WebSocket", "errors", err.Error())

			default:
				statusCode = http.StatusInternalServerError
				message = "Internal server error"
				log.Error("Internal server error")
			}

			c.JSON(statusCode, gin.H{
				"message": message,
			})
		}
	}
}
