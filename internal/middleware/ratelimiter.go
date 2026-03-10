package middleware

import (
	"log/slog"
	"net"
	"net/http"
	"strings"

	errs "github.com/kkonst40/ichat/internal/domain/errors"
	"github.com/kkonst40/ichat/internal/handler"
	"github.com/kkonst40/ichat/internal/ratelimiter"
)

func LimitRate(limiter *ratelimiter.IPRateLimiter) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := getRealIP(r)
			if !limiter.GetLimiter(ip).Allow() {
				handler.WriteError(r.Context(), w, errs.ErrTooManyRequests)
				return
			}
			slog.InfoContext(r.Context(), "", "IP", ip)
			next.ServeHTTP(w, r)
		})
	}
}

func getRealIP(r *http.Request) string {
	xForwardedFor := r.Header.Get("X-Forwarded-For")
	if xForwardedFor != "" {
		ips := strings.Split(xForwardedFor, ",")
		return strings.TrimSpace(ips[0])
	}

	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return ip
}
