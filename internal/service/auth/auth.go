package auth

import (
	"context"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/kkonst40/chat-service/internal/config"
)

type ctxKey struct{}

var userIDKey ctxKey

func GetUserID(ctx context.Context) uuid.UUID {
	return ctx.Value(userIDKey).(uuid.UUID)
}

func ContextWithUserID(ctx context.Context, userID uuid.UUID) context.Context {
	return context.WithValue(ctx, userIDKey, userID)
}

type UserClaims struct {
	ID        uuid.UUID `json:"id"`
	UserName  string    `json:"userName"`
	SessionID uuid.UUID `json:"sid"`
	jwt.RegisteredClaims
}

type TokenValidator struct {
	secretKey []byte
	issuer    string
	audience  string
}

func NewTokenValidator(cfg *config.Config) *TokenValidator {
	return &TokenValidator{
		secretKey: []byte(cfg.JWT.SecretKey),
		issuer:    cfg.JWT.Issuer,
		audience:  cfg.JWT.Audience,
	}
}

func (v *TokenValidator) ValidateToken(tokenString string) (uuid.UUID, error) {
	claims := &UserClaims{}

	token, err := jwt.ParseWithClaims(
		tokenString,
		claims,
		func(token *jwt.Token) (any, error) {
			if token.Method != jwt.SigningMethodHS256 {
				return nil, jwt.ErrSignatureInvalid
			}
			return v.secretKey, nil
		},
		jwt.WithIssuer(v.issuer),
		jwt.WithAudience(v.audience),
	)

	if err != nil {
		return uuid.Nil, err
	}

	if !token.Valid {
		return uuid.Nil, jwt.ErrTokenInvalidClaims
	}

	return claims.ID, nil
}
