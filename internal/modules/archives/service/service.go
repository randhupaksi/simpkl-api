package service

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"gorm.io/gorm"

	"simpkl-api/internal/modules/archives/entity"
	apperrors "simpkl-api/internal/shared/errors"
)

type Service struct{ db *gorm.DB }

func New(db *gorm.DB) *Service { return &Service{db} }

func (s *Service) Archive(
	ctx context.Context,
	periodID, actorID, reason string,
) (*entity.Archive, error) {
	var period struct {
		ID     string
		Name   string
		Status string
	}
	if err := s.db.WithContext(ctx).Table("periods").Where("id = ? AND deleted_at IS NULL", periodID).Take(&period).Error; err != nil {
		return nil, &apperrors.AppError{Status: http.StatusNotFound, Code: "PERIOD_NOT_FOUND", Message: "Periode tidak ditemukan", Cause: err}
	}
	if period.Status != "completed" {
		return nil, &apperrors.AppError{Status: http.StatusConflict, Code: "PERIOD_NOT_COMPLETED", Message: "Periode harus berstatus selesai sebelum diarsipkan"}
	}
	var activePlacements int64
	if err := s.db.WithContext(ctx).Table("placements").
		Where("period_id = ? AND status IN ? AND deleted_at IS NULL", periodID, []string{"draft", "pending_verification", "approved", "ready", "active"}).
		Count(&activePlacements).Error; err != nil {
		return nil, err
	}
	if activePlacements > 0 {
		return nil, &apperrors.AppError{Status: http.StatusConflict, Code: "ACTIVE_PLACEMENTS_EXIST", Message: "Masih terdapat penempatan yang belum diselesaikan"}
	}
	var participantCount, companyCount, documentCount int64
	_ = s.db.WithContext(ctx).Table("placements").Where("period_id = ? AND deleted_at IS NULL", periodID).Distinct("student_id").Count(&participantCount).Error
	_ = s.db.WithContext(ctx).Table("placements").Where("period_id = ? AND deleted_at IS NULL", periodID).Distinct("company_id").Count(&companyCount).Error
	_ = s.db.WithContext(ctx).Table("documents").Where("period_id = ? AND deleted_at IS NULL", periodID).Count(&documentCount).Error
	snapshot, _ := json.Marshal(map[string]any{
		"period_name": period.Name, "participants": participantCount,
		"companies": companyCount, "documents": documentCount, "archived_at": time.Now(),
	})
	archive := &entity.Archive{
		PeriodID: periodID, ArchivedBy: actorID, ArchivedAt: time.Now(),
		Reason: reason, Snapshot: string(snapshot),
	}
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(archive).Error; err != nil {
			return err
		}
		return tx.Table("periods").Where("id = ?", periodID).Update("status", "archived").Error
	})
	return archive, err
}
