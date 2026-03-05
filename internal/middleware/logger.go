package middleware

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/google/uuid"
)

func Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID, _ := uuid.NewV7()
		ctx := context.WithValue(r.Context(), "requestID", requestID.String())

		slog.InfoContext(
			ctx,
			"Request started",
			"method", r.Method,
			"path", r.URL.Path,
		)

		start := time.Now()
		next.ServeHTTP(w, r.WithContext(ctx))
		slog.InfoContext(ctx, "Request handling time", "time", time.Since(start))
	})
}
