package logger

import (
	"context"
	"log/slog"
	"os"

	"github.com/google/uuid"
)

type ctxKey struct{}

var requestIDKey = ctxKey{}

func ContextWithRequestID(ctx context.Context, requestID uuid.UUID) context.Context {
	return context.WithValue(ctx, requestIDKey, requestID.String())
}

type ContextHandler struct {
	slog.Handler
}

func (h *ContextHandler) Handle(ctx context.Context, r slog.Record) error {
	if id, ok := ctx.Value(requestIDKey).(string); ok {
		r.AddAttrs(slog.String("requestID", id))
	}
	return h.Handler.Handle(ctx, r)
}

func New(env string) *slog.Logger {
	var baseHandler slog.Handler
	switch env {
	case "dev":
		baseHandler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug})
	case "prod":
		baseHandler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})
	default:
		baseHandler = slog.Default().Handler()
	}

	logger := slog.New(&ContextHandler{baseHandler})

	return logger
}
