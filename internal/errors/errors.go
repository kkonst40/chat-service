package errors

import (
	"context"
	"errors"
	"log/slog"
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
	ErrChatConnection  = errors.New("chat connection error")
	ErrExternalService = errors.New("external service error")

	ErrInvalidAction = errors.New("unknown action")
)

// should be renamed
type InternalError struct {
	Msg string
}

func (e *InternalError) Error() string {
	return e.Msg
}

type NotFoundError struct {
	Msg string
}

func (e *NotFoundError) Error() string {
	return e.Msg
}

type ForbiddenError struct {
	Msg string
}

func (e *ForbiddenError) Error() string {
	return e.Msg
}

type InvalidRequestError struct {
	Msg string
}

func (e *InvalidRequestError) Error() string {
	return e.Msg
}

type UnauthorizedError struct {
	Msg string
}

func (e *UnauthorizedError) Error() string {
	return e.Msg
}

type DBError struct {
	Msg string
}

func (e *DBError) Error() string {
	return e.Msg
}

type ChatConnectionError struct {
	Msg string
}

func (e *ChatConnectionError) Error() string {
	return e.Msg
}

type ExternalServiceError struct {
	Msg string
}

func (e *ExternalServiceError) Error() string {
	return e.Msg
}

var _ error = (*InternalError)(nil)
var _ error = (*NotFoundError)(nil)
var _ error = (*ForbiddenError)(nil)
var _ error = (*InvalidRequestError)(nil)
var _ error = (*UnauthorizedError)(nil)
var _ error = (*DBError)(nil)
var _ error = (*ChatConnectionError)(nil)
var _ error = (*ExternalServiceError)(nil)

type ErrResp struct {
	Message string `json:"message"`
}

func MapError_(err error, log *slog.Logger) (int, ErrResp) {
	if err == nil {
		return 0, ErrResp{}
	}

	var (
		statusCode int
		msg        string

		internalErr *InternalError
		nfErr       *NotFoundError
		invReqErr   *InvalidRequestError
		frbErr      *ForbiddenError
		unauthErr   *UnauthorizedError
		dbErr       *DBError
		chatConnErr *ChatConnectionError
	)

	switch {
	case errors.As(err, &internalErr):
		log.Error("Internal server error", "error", err)
		statusCode = http.StatusInternalServerError
		msg = "Internal server error"

	case errors.Is(err, context.DeadlineExceeded):
		log.Error("Context deadline exceeded", "error", err)
		statusCode = http.StatusGatewayTimeout
		msg = "The request took too long to process"

	case errors.As(err, &nfErr):
		log.Error("Resource not found", "error", err)
		statusCode = http.StatusNotFound
		msg = nfErr.Msg

	case errors.As(err, &invReqErr):
		log.Error("Invalid request", "error", err)
		statusCode = http.StatusBadRequest
		msg = invReqErr.Msg

	case errors.As(err, &frbErr):
		log.Error("Access denied", "error", err)
		statusCode = http.StatusForbidden
		msg = "Access denied"

	case errors.As(err, &unauthErr):
		log.Error("Unauthorized", "error", err)
		statusCode = http.StatusUnauthorized
		msg = "User unauthorized"

	case errors.As(err, &dbErr):
		log.Error("DB error", "error", err)
		statusCode = http.StatusInternalServerError
		msg = "Internal server error"

	case errors.As(err, &chatConnErr):
		log.Error("WebSocket upgrade error", "error", err)
		statusCode = http.StatusForbidden
		msg = "Upgrading to WebSocket error"

	default:
		log.Error("Unhandled error", "error", err)
		statusCode = http.StatusInternalServerError
		msg = "Internal server error"
	}

	return statusCode, ErrResp{msg}
}

func MapError(err error, log *slog.Logger) (int, ErrResp) {
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

	default:
		statusCode = http.StatusInternalServerError
		msg = "Internal server error"
	}

	log.Error(err.Error())

	return statusCode, ErrResp{msg}
}
