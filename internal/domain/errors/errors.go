package errors

import (
	"context"
	"errors"
	"net/http"
)

var (
	ErrInternal        = errors.New("internal error")
	ErrChatNotFound    = errors.New("chat not found")
	ErrUserNotFound    = errors.New("user not found")
	ErrMsgNotFound     = errors.New("message not found")
	ErrInvalidRequest  = errors.New("invalid request")
	ErrForbidden       = errors.New("access forbidden")
	ErrUnauthorized    = errors.New("unauthorized")
	ErrDatabase        = errors.New("DB error")
	ErrExternalService = errors.New("external service error")

	ErrTooManyRequests        = errors.New("too many requests")
	ErrChatConnection         = errors.New("chat connection error")
	ErrTooManyOpenConnections = errors.New("too many open connections")

	ErrInvalidAction = errors.New("unknown action")
)

type ErrResp struct {
	Message string `json:"message"`
}

func MapError(err error) (int, ErrResp) {
	if err == nil {
		return 0, ErrResp{}
	}

	var (
		statusCode int
		msg        string
	)

	switch {
	case errors.Is(err, context.DeadlineExceeded):
		statusCode = http.StatusGatewayTimeout
		msg = "The request took too long to process"

	case errors.Is(err, ErrInternal):
		statusCode = http.StatusInternalServerError
		msg = "Internal server error"

	case errors.Is(err, ErrChatNotFound) || errors.Is(err, ErrUserNotFound) || errors.Is(err, ErrMsgNotFound):
		statusCode = http.StatusNotFound
		msg = err.Error()

	case errors.Is(err, ErrInvalidRequest):
		statusCode = http.StatusBadRequest
		msg = err.Error()

	case errors.Is(err, ErrForbidden):
		statusCode = http.StatusForbidden
		msg = "Access denied"

	case errors.Is(err, ErrUnauthorized):
		statusCode = http.StatusUnauthorized
		msg = "User unauthorized"

	case errors.Is(err, ErrDatabase):
		statusCode = http.StatusInternalServerError
		msg = "Internal server error"

	case errors.Is(err, ErrChatConnection):
		statusCode = http.StatusForbidden
		msg = "Chat connection error"

	case errors.Is(err, ErrTooManyRequests):
		statusCode = http.StatusTooManyRequests
		msg = "Too many requests"

	case errors.Is(err, ErrTooManyOpenConnections):
		statusCode = http.StatusTooManyRequests
		msg = "Too many active connections from client IP"

	default:
		statusCode = http.StatusInternalServerError
		msg = "Internal server error"
	}

	return statusCode, ErrResp{msg}
}
