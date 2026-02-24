package handler

import (
	"errors"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
	errs "github.com/kkonst40/ichat/internal/errors"
)

func NewValidator() *validator.Validate {
	v := validator.New()

	v.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})

	return v
}

func handleValidationErr(err error) *errs.InvalidRequestError {
	var ve validator.ValidationErrors
	if errors.As(err, &ve) {
		fields := make([]string, 0, len(ve))
		for _, fe := range ve {
			fields = append(fields, fe.Field())
		}

		return &errs.InvalidRequestError{
			Msg: "Invalid fields in request body: " + strings.Join(fields, ", "),
		}
	}

	return &errs.InvalidRequestError{
		Msg: "Invalid request body",
	}
}
