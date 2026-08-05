package service

import (
	"context"
	"net/http"

	"gorm.io/gorm"

	"simpkl-api/internal/modules/readiness/entity"
	apperrors "simpkl-api/internal/shared/errors"
)

type Service struct{ db *gorm.DB }

func New(db *gorm.DB) *Service { return &Service{db} }

func (s *Service) Recalculate(
	ctx context.Context,
	studentID, periodID string,
) (*entity.Readiness, error) {
	var student struct {
		Name        string
		ClassID     string
		MajorID     string
		Phone       string
		ParentName  string
		ParentPhone string
	}
	if err := s.db.WithContext(ctx).Table("students").Where("id = ? AND deleted_at IS NULL", studentID).Take(&student).Error; err != nil {
		return nil, notFound("STUDENT_NOT_FOUND", "Siswa tidak ditemukan", err)
	}

	var placement struct {
		ID               string
		CompanyID        string
		CompanyContactID string
		SupervisorID     string
		StartDate        any
		EndDate          any
		Status           string
	}
	placementErr := s.db.WithContext(ctx).Table("placements").
		Where("student_id = ? AND period_id = ? AND status NOT IN ? AND deleted_at IS NULL", studentID, periodID, []string{"cancelled", "transferred"}).
		Order("created_at DESC").
		Take(&placement).Error

	check := &entity.Readiness{
		StudentID: studentID, PeriodID: periodID, RequiredCount: 8,
		DataComplete: student.Name != "" && student.ClassID != "" && student.MajorID != "" &&
			student.Phone != "" && student.ParentName != "" && student.ParentPhone != "",
	}
	if placementErr == nil {
		check.PlacementID = placement.ID
		check.CompanyAssigned = placement.CompanyID != ""
		check.ContactAvailable = placement.CompanyContactID != ""
		check.SupervisorAssigned = placement.SupervisorID != ""
		check.DatesSet = placement.StartDate != nil && placement.EndDate != nil
	}

	check.AcceptanceLetterValid = s.hasValidDocument(ctx, studentID, periodID, "acceptance_letter")
	check.ParentPermissionValid = s.hasValidDocument(ctx, studentID, periodID, "parent_permission")
	check.IntroductionLetterValid = s.hasValidDocument(ctx, studentID, periodID, "introduction_letter")
	flags := []bool{
		check.DataComplete, check.CompanyAssigned, check.ContactAvailable,
		check.SupervisorAssigned, check.DatesSet, check.AcceptanceLetterValid,
		check.ParentPermissionValid, check.IntroductionLetterValid,
	}
	for _, complete := range flags {
		if complete {
			check.CompletedCount++
		}
	}
	check.Percentage = float64(check.CompletedCount) / float64(check.RequiredCount) * 100
	if check.CompletedCount == check.RequiredCount {
		check.Status = "ready"
	} else if check.CompletedCount >= 5 {
		check.Status = "attention"
	} else {
		check.Status = "incomplete"
	}

	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var existing entity.Readiness
		err := tx.Where("student_id = ? AND period_id = ?", studentID, periodID).First(&existing).Error
		if err == nil {
			check.ID = existing.ID
			check.CreatedAt = existing.CreatedAt
			if err := tx.Model(&existing).Select("*").Omit("id", "created_at", "deleted_at", "PlacementID").Updates(check).Error; err != nil {
				return err
			}
			if err := savePlacementReference(tx, existing.ID, check.PlacementID); err != nil {
				return err
			}
		} else if err == gorm.ErrRecordNotFound {
			if err := tx.Omit("PlacementID").Create(check).Error; err != nil {
				return err
			}
			if err := savePlacementReference(tx, check.ID, check.PlacementID); err != nil {
				return err
			}
		} else {
			return err
		}
		if check.Status == "ready" && placement.ID != "" {
			if err := tx.Table("placements").Where("id = ? AND status = ?", placement.ID, "approved").Update("status", "ready").Error; err != nil {
				return err
			}
			return tx.Table("students").Where("id = ?", studentID).Update("pkl_status", "ready").Error
		}
		return nil
	})
	return check, err
}

func savePlacementReference(tx *gorm.DB, readinessID, placementID string) error {
	var value any
	if placementID != "" {
		value = placementID
	}
	return tx.Model(&entity.Readiness{}).
		Where("id = ?", readinessID).
		Update("placement_id", value).Error
}

func (s *Service) Override(
	ctx context.Context,
	studentID, periodID, reason string,
) (*entity.Readiness, error) {
	if reason == "" {
		return nil, &apperrors.AppError{Status: http.StatusUnprocessableEntity, Code: "OVERRIDE_REASON_REQUIRED", Message: "Alasan pengecualian wajib diisi"}
	}
	check, err := s.Recalculate(ctx, studentID, periodID)
	if err != nil {
		return nil, err
	}
	check.Status = "ready"
	check.OverrideReason = reason
	check.Percentage = 100
	if err := s.db.WithContext(ctx).Model(check).Updates(map[string]any{
		"status": "ready", "override_reason": reason, "percentage": 100,
	}).Error; err != nil {
		return nil, err
	}
	if check.PlacementID != "" {
		_ = s.db.WithContext(ctx).Table("placements").Where("id = ?", check.PlacementID).Update("status", "ready").Error
	}
	_ = s.db.WithContext(ctx).Table("students").Where("id = ?", studentID).Update("pkl_status", "ready").Error
	return check, nil
}

func (s *Service) hasValidDocument(ctx context.Context, studentID, periodID, code string) bool {
	var count int64
	_ = s.db.WithContext(ctx).Table("documents").
		Joins("JOIN document_types ON document_types.id = documents.document_type_id").
		Where(
			"documents.owner_type = ? AND documents.owner_id = ? AND documents.period_id = ? AND documents.status = ? AND document_types.code = ? AND documents.deleted_at IS NULL",
			"student", studentID, periodID, "valid", code,
		).
		Count(&count).Error
	return count > 0
}

func notFound(code, message string, cause error) error {
	return &apperrors.AppError{Status: http.StatusNotFound, Code: code, Message: message, Cause: cause}
}
