package permission

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"

	"simpkl-api/internal/shared/response"
)

type Checker interface {
	HasPermission(context.Context, string, string) (bool, error)
	LoadScope(context.Context, string) ([]string, string, string, error)
}

func New(checker Checker) func(string) gin.HandlerFunc {
	return func(code string) gin.HandlerFunc {
		return func(c *gin.Context) {
			userID := c.GetString("user_id")
			allowed, err := checker.HasPermission(c.Request.Context(), userID, code)
			if err != nil {
				response.InternalError(c)
				return
			}
			if !allowed {
				response.Error(c, http.StatusForbidden, "Anda tidak memiliki izin untuk tindakan ini", "FORBIDDEN", nil)
				return
			}
			roles, majorID, classID, err := checker.LoadScope(c.Request.Context(), userID)
			if err != nil {
				response.InternalError(c)
				return
			}
			c.Set("user_roles", roles)
			if scopedRole(roles, "program_head") && majorID != "" {
				c.Set("scope_major_id", majorID)
			}
			if scopedRole(roles, "homeroom_teacher") && classID != "" {
				c.Set("scope_class_id", classID)
			}
			c.Next()
		}
	}
}

func scopedRole(roles []string, target string) bool {
	for _, role := range roles {
		if role == "super_admin" || role == "admin_pkl" || role == "coordinator_pkl" {
			return false
		}
		if role == target {
			return true
		}
	}
	return false
}
