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

type DBError struct {
	Msg string
}

func (e *DBError) Error() string {
	return e.Msg
}

var _ error = (*NotFoundError)(nil)
var _ error = (*ForbiddenError)(nil)
var _ error = (*DBError)(nil)
