package users

import (
	"context"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	userhttp "simpkl-api/internal/modules/users/delivery/http"
	"simpkl-api/internal/modules/users/entity"
	userservice "simpkl-api/internal/modules/users/service"
	platformauth "simpkl-api/internal/platform/auth"
	"simpkl-api/internal/shared/crud"
	apperrors "simpkl-api/internal/shared/errors"
	"simpkl-api/internal/shared/types"
)

func Register(api *gin.RouterGroup, db *gorm.DB, auditor types.Auditor, require func(string) gin.HandlerFunc) {
	repo := crud.NewGormRepository[entity.User](
		db,
		[]string{"name", "email", "username"},
		map[string]string{"major_id": "major_id", "class_id": "class_id", "status": "status"},
	)
	validate := func(_ context.Context, existing *entity.User, user *entity.User) error {
		user.Email = strings.ToLower(strings.TrimSpace(user.Email))
		user.Username = strings.ToLower(strings.TrimSpace(user.Username))
		if user.Password != "" {
			hash, err := platformauth.HashPassword(user.Password)
			if err != nil {
				return err
			}
			user.PasswordHash = hash
			user.Password = ""
		} else if existing != nil {
			user.PasswordHash = existing.PasswordHash
		} else {
			return &apperrors.AppError{Status: http.StatusUnprocessableEntity, Code: "PASSWORD_REQUIRED", Message: "Password wajib diisi untuk pengguna baru"}
		}
		return nil
	}
	service := crud.NewService("user", repo, auditor, validate, nil)
	handler := crud.NewHandler(service, "major_id", "class_id", "status")
	group := api.Group("/users")
	crud.RegisterRoutes(group, handler, require, "user")
	assignments := userhttp.NewHandler(userservice.NewRoleAssignmentService(db))
	group.PUT("/:id/roles", require("user.manage"), assignments.SetRoles)
}
