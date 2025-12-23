package apperror

import "errors"

var (
	ErrUserNotFound         = errors.New("user not found")
	ErrChatNotFound         = errors.New("chat not found")
	ErrChatAlreadyExists    = errors.New("chat already exists")
	ErrMessageNotFound      = errors.New("message not found")
	ErrMessageAlreadyExists = errors.New("message already exists")
)
