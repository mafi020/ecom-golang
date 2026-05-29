package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/mafi020/ecom-golang/config"
	"github.com/mafi020/ecom-golang/internal/apperrors"
	"github.com/mafi020/ecom-golang/internal/delivery/http/utils"
)

func AuthMiddleware(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. Get token from header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			utils.HandleError(c, &apperrors.UnauthorizedError{Message: "missing or invalid token"})
			return
		}

		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")

		// 2. Parse and validate token
		token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (any, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, &apperrors.UnauthorizedError{Message: "unexpected signing method"}
			}
			return []byte(cfg.JWT.Secret), nil
		})

		if err != nil || !token.Valid {
			utils.HandleError(c, &apperrors.UnauthorizedError{Message: "invalid or expired token"})
			return
		}

		// 3. Set claims in context
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			utils.HandleError(c, &apperrors.UnauthorizedError{Message: "invalid token claims"})
			return
		}

		c.Set("user_id", int64(claims["sub"].(float64)))
		c.Set("role", claims["role"].(string))

		c.Next()
	}
}
