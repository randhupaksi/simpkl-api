package service

import (
	"context"
	"net/http"

	"gorm.io/gorm"

	"simpkl-api/internal/modules/companies/entity"
	apperrors "simpkl-api/internal/shared/errors"
	"simpkl-api/internal/shared/pagination"
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

// EligibleCompanies returns companies that can accept the selected student's major.
// A company without configured major capacities is treated as accepting every major,
// matching placement validation rules.
func (s *PartnershipService) EligibleCompanies(
	ctx context.Context,
	query pagination.Query,
	studentID string,
	selectedCompanyID string,
) ([]entity.Company, pagination.Meta, error) {
	query.Normalize()
	statement := s.db.WithContext(ctx).Model(&entity.Company{})

	if studentID != "" {
		var studentMajorID string
		result := s.db.WithContext(ctx).
			Table("students").
			Where("id = ? AND deleted_at IS NULL", studentID).
			Pluck("major_id", &studentMajorID)
		if result.Error != nil {
			return nil, pagination.Meta{}, result.Error
		}
		if result.RowsAffected == 0 {
			return nil, pagination.Meta{}, &apperrors.AppError{
				Status:  http.StatusUnprocessableEntity,
				Code:    "STUDENT_NOT_FOUND",
				Message: "Siswa yang dipilih tidak ditemukan. Pilih siswa lain lalu coba lagi.",
				Errors:  map[string][]string{"student_id": {"Siswa yang dipilih sudah tidak tersedia."}},
			}
		}

		if selectedCompanyID != "" {
			statement = statement.Where(`
				companies.id = ? OR NOT EXISTS (
					SELECT 1 FROM company_major_capacities configured
					WHERE configured.company_id = companies.id
				) OR EXISTS (
					SELECT 1 FROM company_major_capacities accepted
					WHERE accepted.company_id = companies.id AND accepted.major_id = ?
				)
			`, selectedCompanyID, studentMajorID)
		} else {
			statement = statement.Where(`
				NOT EXISTS (
					SELECT 1 FROM company_major_capacities configured
					WHERE configured.company_id = companies.id
				) OR EXISTS (
					SELECT 1 FROM company_major_capacities accepted
					WHERE accepted.company_id = companies.id AND accepted.major_id = ?
				)
			`, studentMajorID)
		}
	}

	if query.Search != "" {
		statement = statement.Where("name LIKE ? OR industry LIKE ? OR city LIKE ?", "%"+query.Search+"%", "%"+query.Search+"%", "%"+query.Search+"%")
	}

	var total int64
	if err := statement.Count(&total).Error; err != nil {
		return nil, pagination.Meta{}, err
	}

	var companies []entity.Company
	if err := statement.Order("name ASC").Offset(query.Offset()).Limit(query.PerPage).Find(&companies).Error; err != nil {
		return nil, pagination.Meta{}, err
	}

	return companies, pagination.NewMeta(query, total), nil
}
