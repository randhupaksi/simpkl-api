package service

import (
	"context"
	"net/http"

	"gorm.io/gorm"

	roleentity "simpkl-api/internal/modules/roles/entity"
	apperrors "simpkl-api/internal/shared/errors"
)

type RoleAssignmentService struct{ db *gorm.DB }

func NewRoleAssignmentService(db *gorm.DB) *RoleAssignmentService {
	return &RoleAssignmentService{db}
}

func (s *RoleAssignmentService) SetRoles(
	ctx context.Context,
	userID string,
	roleIDs []string,
) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var userCount int64
		if err := tx.Table("users").Where("id = ? AND deleted_at IS NULL", userID).Count(&userCount).Error; err != nil {
			return err
		}
		if userCount == 0 {
			return &apperrors.AppError{Status: http.StatusNotFound, Code: "USER_NOT_FOUND", Message: "Pengguna tidak ditemukan"}
		}
		if len(roleIDs) > 0 {
			var roleCount int64
			if err := tx.Table("roles").Where("id IN ? AND deleted_at IS NULL", roleIDs).Count(&roleCount).Error; err != nil {
				return err
			}
			if roleCount != int64(len(roleIDs)) {
				return &apperrors.AppError{Status: http.StatusUnprocessableEntity, Code: "INVALID_ROLE", Message: "Salah satu role tidak ditemukan"}
			}
		}
		if err := tx.Where("user_id = ?", userID).Delete(&roleentity.UserRole{}).Error; err != nil {
			return err
		}
		for _, roleID := range roleIDs {
			if err := tx.Create(&roleentity.UserRole{UserID: userID, RoleID: roleID}).Error; err != nil {
				return err
			}
		}
		return nil
	})
}
