package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/mafi020/ecom-golang/internal/apperrors"
	"github.com/mafi020/ecom-golang/internal/delivery/http/utils"
)

func RequireRole(role string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, exists := c.Get("role")
		if !exists || userRole.(string) != role {
			utils.HandleError(c, &apperrors.ForbiddenError{Message: "forbidden"})
			return
		}
		c.Next()
	}
}
