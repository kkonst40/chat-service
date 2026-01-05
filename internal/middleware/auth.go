package middleware

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"github.com/kkonst40/ichat/internal/config"
	"github.com/kkonst40/ichat/internal/logger"
)

type CustomClaims struct {
	ID       string `json:"id"`
	UserName string `json:"userName"`
	TokenID  string `json:"tokenId"`
	jwt.StandardClaims
}

func Auth(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString, err := c.Cookie(cfg.JWT.CookieName)
		if err != nil {
			c.Status(http.StatusUnauthorized)
			c.Abort()
			return
		}

		token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
			}
			return []byte(cfg.JWT.SecretKey), nil
		})

		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		if claims, ok := token.Claims.(*CustomClaims); ok && token.Valid {
			if claims.Issuer != cfg.JWT.Issuer &&
				claims.Audience != cfg.JWT.Audience &&
				claims.ExpiresAt > time.Now().Unix() {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
				c.Abort()
				return
			}

			c.Set("requesterID", claims.ID)

			ctx := c.Request.Context()
			logger.FromContext(ctx).Info("", "userID", claims.ID)

			c.Next()
		} else {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}
	}
}
