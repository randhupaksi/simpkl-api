package service

import (
	"context"
	"net/http"

	"gorm.io/gorm"

	"simpkl-api/internal/modules/companies/entity"
	apperrors "simpkl-api/internal/shared/errors"
)

type PartnershipService struct{ db *gorm.DB }

func NewPartnershipService(db *gorm.DB) *PartnershipService {
	return &PartnershipService{db}
}

func (s *PartnershipService) SetMajorCapacities(
	ctx context.Context,
	companyID string,
	capacities []entity.MajorCapacity,
) ([]entity.MajorCapacity, error) {
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var companyCount int64
		if err := tx.Table("companies").Where("id = ? AND deleted_at IS NULL", companyID).Count(&companyCount).Error; err != nil {
			return err
		}
		if companyCount == 0 {
			return &apperrors.AppError{Status: http.StatusNotFound, Code: "COMPANY_NOT_FOUND", Message: "Perusahaan tidak ditemukan"}
		}
		if err := tx.Where("company_id = ?", companyID).Delete(&entity.MajorCapacity{}).Error; err != nil {
			return err
		}
		for index := range capacities {
			capacities[index].CompanyID = companyID
			var majorCount int64
			if err := tx.Table("majors").Where("id = ? AND status = ? AND deleted_at IS NULL", capacities[index].MajorID, "active").Count(&majorCount).Error; err != nil {
				return err
			}
			if majorCount == 0 {
				return &apperrors.AppError{Status: http.StatusUnprocessableEntity, Code: "MAJOR_NOT_FOUND", Message: "Salah satu jurusan tidak ditemukan atau tidak aktif"}
			}
			if err := tx.Create(&capacities[index]).Error; err != nil {
				return err
			}
		}
		return nil
	})
	return capacities, err
}

func (s *PartnershipService) MajorCapacities(
	ctx context.Context,
	companyID string,
) ([]entity.MajorCapacity, error) {
	var result []entity.MajorCapacity
	return result, s.db.WithContext(ctx).Where("company_id = ?", companyID).Find(&result).Error
}
