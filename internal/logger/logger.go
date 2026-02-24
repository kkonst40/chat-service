package logger

import (
	"context"
	"log/slog"
	"os"
)

type ContextKey string

const (
	LoggerCtxKey ContextKey = "logger"
)

func FromContext(ctx context.Context) *slog.Logger {
	if logger, ok := ctx.Value(LoggerCtxKey).(*slog.Logger); ok {
		return logger
	}
	return slog.Default()
}

func New(env string) *slog.Logger {
	var log *slog.Logger
	switch env {
	case "dev":
		log = slog.New(
			slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case "prod":
		log = slog.New(
			slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
		)
	}

	return log
}
