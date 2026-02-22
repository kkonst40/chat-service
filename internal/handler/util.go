package handler

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/kkonst40/ichat/internal/apperror"
)

func WriteJSON(w http.ResponseWriter, statusCode int, body any, log *slog.Logger) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(body); err != nil {
		log.Error("JSON encoding error")
	}
}

func WriteString(w http.ResponseWriter, statusCode int, err string, log *slog.Logger) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	body := apperror.ErrResp{Message: err}
	if err := json.NewEncoder(w).Encode(body); err != nil {
		log.Error("JSON encoding error")
	}
}

func bindJSON(r *http.Request, dst any, validate *validator.Validate) error {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(dst); err != nil {
		return err
	}

	return validate.Struct(dst)
}

type ContextKey string

const (
	UserIDCtxKey ContextKey = "userID"
)

func getUserID(ctx context.Context) uuid.UUID {
	userID := ctx.Value(UserIDCtxKey).(string)
	return uuid.MustParse(userID)
}
