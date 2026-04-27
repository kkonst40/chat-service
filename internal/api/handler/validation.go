package handler

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
	errs "github.com/kkonst40/chat-service/internal/domain/errors"
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

func handleValidationErr(err error) error {
	var ve validator.ValidationErrors
	if errors.As(err, &ve) {
		fields := make([]string, 0, len(ve))
		for _, fe := range ve {
			fields = append(fields, fe.Field())
		}

		return fmt.Errorf(
			"%w: fields: %s",
			errs.ErrInvalidRequest,
			strings.Join(fields, ", "),
		)
	}

	return fmt.Errorf("%w: body", errs.ErrInvalidRequest)
}
