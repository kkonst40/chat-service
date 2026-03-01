package middleware

import (
	"context"
	"fmt"
	"net/http"

	"github.com/golang-jwt/jwt/v5"

	"github.com/google/uuid"
	"github.com/kkonst40/ichat/internal/config"
	errs "github.com/kkonst40/ichat/internal/domain/errors"
	"github.com/kkonst40/ichat/internal/handler"
	"github.com/kkonst40/ichat/internal/logger"
)

type UserClaims struct {
	ID       uuid.UUID `json:"id"`
	UserName string    `json:"userName"`
	TokenID  uuid.UUID `json:"tokenId"`
	jwt.RegisteredClaims
}

func Auth(cfg *config.Config) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			log := logger.FromContext(ctx)

			token, err := r.Cookie(cfg.JWT.CookieName)
			if err != nil {
				log.Error("Token not found", "error", err.Error())
				handler.WriteError(w, fmt.Errorf("%w: invalid token", errs.ErrUnauthorized), log)
				return
			}

			log.Debug("", "token", token.Value)

			claims, err := validateToken(token.Value, cfg)
			if err != nil {
				log.Error("Token validation error", "error", err.Error())
				handler.WriteError(w, fmt.Errorf("%w: invalid token", errs.ErrUnauthorized), log)
				return
			}

			log.Debug("", "userID", claims.ID)

			ctx = context.WithValue(ctx, "requesterID", claims.ID.String())

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func DummyAuthQ(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := r.URL.Query().Get("userId")
		ctx := context.WithValue(r.Context(), "requesterID", userID)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func DummyAuthH(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := r.Header.Get("UserID")
		ctx := context.WithValue(r.Context(), "requesterID", userID)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func validateToken(tokenString string, cfg *config.Config) (*UserClaims, error) {
	claims := &UserClaims{}

	token, err := jwt.ParseWithClaims(
		tokenString,
		claims,
		func(token *jwt.Token) (any, error) {
			if token.Method != jwt.SigningMethodHS256 {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(cfg.JWT.SecretKey), nil
		},
		jwt.WithIssuer(cfg.JWT.Issuer),
		jwt.WithAudience(cfg.JWT.Audience),
	)

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, jwt.ErrTokenInvalidClaims
	}

	return claims, nil
}
