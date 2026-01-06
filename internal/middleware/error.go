package middleware

import (
	"context"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kkonst40/ichat/internal/apperror"
)

func Error() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if len(c.Errors) > 0 {
			err := c.Errors.Last()

			var statusCode int
			var message string

			var (
				nfErr     *apperror.NotFoundError
				invReqErr *apperror.InvalidRequestError
				frbErr    *apperror.ForbiddenError
				unauthErr *apperror.UnauthorizedError
				dbErr     *apperror.DBError
			)

			switch {
			case errors.Is(err, context.DeadlineExceeded):
				statusCode = http.StatusGatewayTimeout
				message = "The request took too long to process"
			case errors.As(err, &nfErr):
				statusCode = http.StatusNotFound
				message = nfErr.Msg
			case errors.As(err, &invReqErr):
				statusCode = http.StatusBadRequest
				message = invReqErr.Msg
			case errors.As(err, &frbErr):
				statusCode = http.StatusForbidden
				message = "Access denied"
			case errors.As(err, &unauthErr):
				statusCode = http.StatusUnauthorized
				message = "User unauthorized"
			case errors.As(err, &dbErr):
				statusCode = http.StatusInternalServerError
				message = "Internal server error"
			default:
				statusCode = http.StatusInternalServerError
				message = "Internal server error"
			}

			c.JSON(statusCode, gin.H{
				"message": message,
			})
		}
	}
}
