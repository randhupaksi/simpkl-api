package auth

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	platformauth "simpkl-api/internal/platform/auth"
	"simpkl-api/internal/shared/response"
)

func New(tokens *platformauth.TokenManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if !strings.HasPrefix(header, "Bearer ") {
			response.Error(c, http.StatusUnauthorized, "Token akses diperlukan", "UNAUTHORIZED", nil)
			return
		}
		claims, err := tokens.ParseAccess(strings.TrimSpace(strings.TrimPrefix(header, "Bearer ")))
		if err != nil {
			response.Error(c, http.StatusUnauthorized, "Token akses tidak valid atau kedaluwarsa", "UNAUTHORIZED", nil)
			return
		}
		c.Set("user_id", claims.UserID)
		c.Set("user_email", claims.Email)
		c.Next()
	}
}
