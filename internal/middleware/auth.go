package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"

	"github.com/google/uuid"
	"github.com/kkonst40/ichat/internal/apperror"
	"github.com/kkonst40/ichat/internal/config"
	"github.com/kkonst40/ichat/internal/logger"
)

type UserClaims struct {
	ID       uuid.UUID `json:"id"`
	UserName string    `json:"userName"`
	TokenID  uuid.UUID `json:"tokenId"`
	jwt.RegisteredClaims
}

func Auth(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		log := logger.FromContext(ctx)

		tokenString, err := c.Cookie(cfg.JWT.CookieName)
		if err != nil {
			c.Error(&apperror.UnauthorizedError{Msg: "Token not found"})
			c.Abort()
			return
		}

		log.Info("", "token", tokenString)

		claims, err := validateToken(tokenString, cfg)
		if err != nil {
			log.Error("Token validation error", "error", err.Error())
			c.Error(&apperror.UnauthorizedError{Msg: "Invalid token"})
			c.Abort()
			return
		}
		c.Set("requesterID", claims.ID.String())

		log.Info("", "userID", claims.ID)

		c.Next()
	}
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
