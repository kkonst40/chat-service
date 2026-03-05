package handler

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	errs "github.com/kkonst40/ichat/internal/domain/errors"
)

func WriteJSON(ctx context.Context, w http.ResponseWriter, statusCode int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(body); err != nil {
		slog.ErrorContext(ctx, "JSON encoding", "error", err)
	}
}

func WriteError(ctx context.Context, w http.ResponseWriter, err error) {
	statusCode, resp := errs.MapError(err)
	slog.ErrorContext(ctx, err.Error())

	WriteJSON(ctx, w, statusCode, resp)
}

func bindJSON(r *http.Request, dst any, validate *validator.Validate) error {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(dst); err != nil {
		return err
	}

	return validate.Struct(dst)
}

func getUserID(ctx context.Context) uuid.UUID {
	userID := ctx.Value("requesterID").(string)
	return uuid.MustParse(userID)
}
