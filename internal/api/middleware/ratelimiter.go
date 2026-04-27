package middleware

import (
	"log/slog"
	"net/http"

	"github.com/kkonst40/chat-service/internal/api/handler"
	"github.com/kkonst40/chat-service/internal/api/limit/ratelimiter"
	errs "github.com/kkonst40/chat-service/internal/domain/errors"
)

func LimitRate(limiter *ratelimiter.IPRateLimiter) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := handler.GetRealIP(r)
			if !limiter.GetLimiter(ip).Allow() {
				handler.WriteError(r.Context(), w, errs.ErrTooManyRequests)
				return
			}
			slog.InfoContext(r.Context(), "", "client IP", ip)
			next.ServeHTTP(w, r)
		})
	}
}
