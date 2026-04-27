package middleware

import (
	"log/slog"
	"net/http"

	"github.com/kkonst40/chat-service/internal/api/handler"
	errs "github.com/kkonst40/chat-service/internal/domain/errors"
	"github.com/kkonst40/chat-service/internal/limit/ratelimiter"
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
