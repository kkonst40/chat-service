package middleware

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/google/uuid"
	"github.com/kkonst40/chat-service/internal/api/handler"
	"github.com/kkonst40/chat-service/internal/auth"
	errs "github.com/kkonst40/chat-service/internal/domain/errors"
)

func Auth(validator *auth.TokenValidator, cookieName string) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			token, err := r.Cookie(cookieName)
			if err != nil {
				handler.WriteError(ctx, w, fmt.Errorf("%w: token not found: %w", errs.ErrUnauthorized, err))
				return
			}

			userID, err := validator.ValidateToken(token.Value)
			if err != nil {
				handler.WriteError(ctx, w, fmt.Errorf("%w: token validation error: %w", errs.ErrUnauthorized, err))
				return
			}

			slog.DebugContext(ctx, "", "userID", userID)

			ctx = auth.ContextWithUserID(ctx, userID)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func DummyAuthQ(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userIDString := r.URL.Query().Get("userId")
		userID, err := uuid.Parse(userIDString)
		if err != nil {
			handler.WriteError(r.Context(), w, errs.ErrUnauthorized)
		}

		ctx := auth.ContextWithUserID(r.Context(), userID)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func DummyAuthH(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userIDString := r.Header.Get("UserID")
		userID, err := uuid.Parse(userIDString)
		if err != nil {
			handler.WriteError(r.Context(), w, errs.ErrUnauthorized)
		}

		ctx := auth.ContextWithUserID(r.Context(), userID)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
