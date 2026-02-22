package middleware

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/kkonst40/ichat/internal/logger"
)

func Logger(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		requestID, _ := uuid.NewV7()
		log := slog.With("requestID", requestID.String())

		log.Info("Request started",
			"method", r.Method,
			"path", r.URL.Path,
		)

		ctx := context.WithValue(r.Context(), logger.LoggerCtxKey, log)

		start := time.Now()
		next(w, r.WithContext(ctx))
		log.Info("Request handling time", "time", time.Since(start))
	}
}
