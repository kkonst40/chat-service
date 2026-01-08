package apperror

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

var _ error = (*NotFoundError)(nil)
var _ error = (*ForbiddenError)(nil)
var _ error = (*InvalidRequestError)(nil)
var _ error = (*UnauthorizedError)(nil)
var _ error = (*DBError)(nil)
var _ error = (*ChatConnectionError)(nil)
