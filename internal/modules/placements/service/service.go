package service

import (
	"context"
	"net/http"
	"time"

	"gorm.io/gorm"

	"simpkl-api/internal/modules/placements/entity"
	apperrors "simpkl-api/internal/shared/errors"
)

type TransferService struct{ db *gorm.DB }

func NewTransferService(db *gorm.DB) *TransferService { return &TransferService{db} }

func (s *TransferService) Transfer(
	ctx context.Context,
	placementID string,
	endDate time.Time,
	reason string,
	next *entity.Placement,
) (*entity.Placement, error) {
	if reason == "" {
		return nil, &apperrors.AppError{Status: http.StatusUnprocessableEntity, Code: "TRANSFER_REASON_REQUIRED", Message: "Alasan perpindahan wajib diisi"}
	}
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var previous entity.Placement
		if err := tx.First(&previous, "id = ?", placementID).Error; err != nil {
			return err
		}
		if previous.Status != "active" && previous.Status != "ready" && previous.Status != "approved" {
			return &apperrors.AppError{Status: http.StatusConflict, Code: "PLACEMENT_NOT_TRANSFERABLE", Message: "Status penempatan tidak dapat dipindahkan"}
		}
		if err := tx.Model(&previous).Updates(map[string]any{
			"status": "transferred", "end_date": endDate, "transfer_reason": reason,
		}).Error; err != nil {
			return err
		}
		next.StudentID = previous.StudentID
		next.PeriodID = previous.PeriodID
		next.PreviousPlacementID = previous.ID
		if next.Status == "" {
			next.Status = "pending_verification"
		}
		if err := tx.Create(next).Error; err != nil {
			return err
		}
		return tx.Table("students").Where("id = ?", previous.StudentID).Update("pkl_status", "placement_process").Error
	})
	return next, err
}
