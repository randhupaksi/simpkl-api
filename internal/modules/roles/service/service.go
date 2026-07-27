package service

import (
	"context"
	"net/http"

	"gorm.io/gorm"

	permissionentity "simpkl-api/internal/modules/permissions/entity"
	apperrors "simpkl-api/internal/shared/errors"
)

type PermissionAssignmentService struct{ db *gorm.DB }

func NewPermissionAssignmentService(db *gorm.DB) *PermissionAssignmentService {
	return &PermissionAssignmentService{db}
}

func (s *PermissionAssignmentService) SetPermissions(
	ctx context.Context,
	roleID string,
	permissionIDs []string,
) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var roleCount int64
		if err := tx.Table("roles").Where("id = ? AND deleted_at IS NULL", roleID).Count(&roleCount).Error; err != nil {
			return err
		}
		if roleCount == 0 {
			return &apperrors.AppError{Status: http.StatusNotFound, Code: "ROLE_NOT_FOUND", Message: "Role tidak ditemukan"}
		}
		if len(permissionIDs) > 0 {
			var permissionCount int64
			if err := tx.Table("permissions").Where("id IN ? AND deleted_at IS NULL", permissionIDs).Count(&permissionCount).Error; err != nil {
				return err
			}
			if permissionCount != int64(len(permissionIDs)) {
				return &apperrors.AppError{Status: http.StatusUnprocessableEntity, Code: "INVALID_PERMISSION", Message: "Salah satu permission tidak ditemukan"}
			}
		}
		if err := tx.Where("role_id = ?", roleID).Delete(&permissionentity.RolePermission{}).Error; err != nil {
			return err
		}
		for _, permissionID := range permissionIDs {
			if err := tx.Create(&permissionentity.RolePermission{RoleID: roleID, PermissionID: permissionID}).Error; err != nil {
				return err
			}
		}
		return nil
	})
}
