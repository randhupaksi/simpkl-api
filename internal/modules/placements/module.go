package placements

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	placementhttp "simpkl-api/internal/modules/placements/delivery/http"
	"simpkl-api/internal/modules/placements/entity"
	placementservice "simpkl-api/internal/modules/placements/service"
	"simpkl-api/internal/shared/crud"
	apperrors "simpkl-api/internal/shared/errors"
	"simpkl-api/internal/shared/types"
)

var placementTransitions = map[string]map[string]bool{
	"draft":                {"draft": true, "pending_verification": true, "cancelled": true},
	"pending_verification": {"pending_verification": true, "approved": true, "draft": true, "cancelled": true},
	"approved":             {"approved": true, "ready": true, "cancelled": true},
	"ready":                {"ready": true, "active": true, "cancelled": true},
	"active":               {"active": true, "completed": true, "transferred": true},
	"completed":            {"completed": true},
	"cancelled":            {"cancelled": true},
	"transferred":          {"transferred": true},
}

func Register(api *gin.RouterGroup, db *gorm.DB, auditor types.Auditor, require func(string) gin.HandlerFunc) {
	repo := crud.NewGormRepository[entity.Placement](
		db,
		[]string{"division", "position", "address", "notes"},
		map[string]string{
			"period_id": "period_id", "student_id": "student_id", "company_id": "company_id",
			"supervisor_id": "supervisor_id", "status": "status", "work_system": "work_system",
		},
	)
	validate := func(ctx context.Context, existing *entity.Placement, placement *entity.Placement) error {
		if !placement.EndDate.After(placement.StartDate) {
			return invalidPlacement("INVALID_DATE_RANGE", "Tanggal selesai harus setelah tanggal mulai")
		}
		if existing != nil && !placementTransitions[existing.Status][placement.Status] {
			return &apperrors.AppError{Status: http.StatusConflict, Code: "INVALID_STATUS_TRANSITION", Message: "Perubahan status penempatan tidak diizinkan"}
		}
		var period struct {
			StartDate time.Time
			EndDate   time.Time
			Status    string
		}
		if err := db.WithContext(ctx).Table("periods").Select("start_date, end_date, status").Where("id = ? AND deleted_at IS NULL", placement.PeriodID).Take(&period).Error; err != nil {
			return invalidPlacement("PERIOD_NOT_FOUND", "Periode PKL tidak ditemukan")
		}
		if period.Status == "completed" || period.Status == "archived" {
			return invalidPlacement("PERIOD_CLOSED", "Periode PKL sudah ditutup")
		}
		if placement.StartDate.Before(period.StartDate) || placement.EndDate.After(period.EndDate) {
			return invalidPlacement("PLACEMENT_OUTSIDE_PERIOD", "Tanggal penempatan harus berada dalam rentang periode PKL")
		}
		var studentCount int64
		if err := db.WithContext(ctx).Table("students").Where("id = ? AND status = ? AND deleted_at IS NULL", placement.StudentID, "active").Count(&studentCount).Error; err != nil || studentCount == 0 {
			return invalidPlacement("STUDENT_NOT_AVAILABLE", "Siswa tidak tersedia")
		}
		var studentMajorID string
		if err := db.WithContext(ctx).Table("students").Where("id = ?", placement.StudentID).Pluck("major_id", &studentMajorID).Error; err != nil {
			return err
		}
		var company struct {
			Status   string
			Capacity int
		}
		if err := db.WithContext(ctx).Table("companies").Select("status, capacity").Where("id = ? AND deleted_at IS NULL", placement.CompanyID).Take(&company).Error; err != nil {
			return invalidPlacement("COMPANY_NOT_FOUND", "Perusahaan tidak ditemukan")
		}
		if company.Status == "blocked" || company.Status == "not_recommended" {
			return invalidPlacement("COMPANY_NOT_ALLOWED", "Perusahaan tidak dapat digunakan untuk penempatan")
		}
		var configuredMajors int64
		if err := db.WithContext(ctx).Table("company_major_capacities").Where("company_id = ?", placement.CompanyID).Count(&configuredMajors).Error; err != nil {
			return err
		}
		var majorCapacity int
		majorQuery := db.WithContext(ctx).Table("company_major_capacities").
			Where("company_id = ? AND major_id = ?", placement.CompanyID, studentMajorID).
			Pluck("capacity", &majorCapacity)
		if err := majorQuery.Error; err != nil {
			return err
		}
		if configuredMajors > 0 && majorQuery.RowsAffected == 0 && placement.OverrideReason == "" {
			return &apperrors.AppError{Status: http.StatusConflict, Code: "MAJOR_NOT_ACCEPTED", Message: "Jurusan siswa belum tercatat diterima perusahaan; alasan pengecualian wajib diisi"}
		}
		if placement.CompanyContactID != "" {
			var contactCount int64
			if err := db.WithContext(ctx).Table("company_contacts").
				Where("id = ? AND company_id = ? AND deleted_at IS NULL", placement.CompanyContactID, placement.CompanyID).
				Count(&contactCount).Error; err != nil || contactCount == 0 {
				return invalidPlacement("INVALID_COMPANY_CONTACT", "PIC tidak terdaftar pada perusahaan yang dipilih")
			}
		}
		if placement.SupervisorID != "" {
			var supervisor struct {
				Status      string
				MaxStudents int
			}
			if err := db.WithContext(ctx).Table("supervisors").
				Select("status, max_students").
				Where("id = ? AND deleted_at IS NULL", placement.SupervisorID).
				Take(&supervisor).Error; err != nil || supervisor.Status != "active" {
				return invalidPlacement("SUPERVISOR_NOT_AVAILABLE", "Guru pembimbing tidak tersedia")
			}
			var supervised int64
			query := db.WithContext(ctx).Table("placements").
				Where("supervisor_id = ? AND period_id = ? AND status IN ? AND deleted_at IS NULL", placement.SupervisorID, placement.PeriodID, []string{"approved", "ready", "active"})
			if existing != nil {
				query = query.Where("id <> ?", existing.ID)
			}
			if err := query.Count(&supervised).Error; err != nil {
				return err
			}
			if supervisor.MaxStudents > 0 && supervised >= int64(supervisor.MaxStudents) && placement.OverrideReason == "" {
				return &apperrors.AppError{Status: http.StatusConflict, Code: "SUPERVISOR_CAPACITY_EXCEEDED", Message: "Beban guru pembimbing sudah penuh; alasan pengecualian wajib diisi"}
			}
		}
		var duplicate int64
		query := db.WithContext(ctx).Table("placements").
			Where("student_id = ? AND period_id = ? AND status IN ? AND deleted_at IS NULL", placement.StudentID, placement.PeriodID, []string{"pending_verification", "approved", "ready", "active"})
		if existing != nil {
			query = query.Where("id <> ?", existing.ID)
		}
		if err := query.Count(&duplicate).Error; err != nil || duplicate > 0 {
			return &apperrors.AppError{Status: http.StatusConflict, Code: "ACTIVE_PLACEMENT_EXISTS", Message: "Siswa sudah memiliki penempatan aktif pada periode ini"}
		}
		if company.Capacity > 0 {
			var used int64
			query := db.WithContext(ctx).Table("placements").
				Where("company_id = ? AND period_id = ? AND status IN ? AND deleted_at IS NULL", placement.CompanyID, placement.PeriodID, []string{"approved", "ready", "active"})
			if existing != nil {
				query = query.Where("id <> ?", existing.ID)
			}
			if err := query.Count(&used).Error; err != nil {
				return err
			}
			if used >= int64(company.Capacity) && placement.OverrideReason == "" {
				return &apperrors.AppError{Status: http.StatusConflict, Code: "COMPANY_CAPACITY_EXCEEDED", Message: "Kuota perusahaan sudah penuh; alasan pengecualian wajib diisi"}
			}
		}
		if majorCapacity > 0 {
			var usedByMajor int64
			query := db.WithContext(ctx).Table("placements").
				Joins("JOIN students ON students.id = placements.student_id").
				Where("placements.company_id = ? AND placements.period_id = ? AND students.major_id = ? AND placements.status IN ? AND placements.deleted_at IS NULL", placement.CompanyID, placement.PeriodID, studentMajorID, []string{"approved", "ready", "active"})
			if existing != nil {
				query = query.Where("placements.id <> ?", existing.ID)
			}
			if err := query.Count(&usedByMajor).Error; err != nil {
				return err
			}
			if usedByMajor >= int64(majorCapacity) && placement.OverrideReason == "" {
				return &apperrors.AppError{Status: http.StatusConflict, Code: "MAJOR_CAPACITY_EXCEEDED", Message: "Kuota perusahaan untuk jurusan siswa sudah penuh; alasan pengecualian wajib diisi"}
			}
		}
		return nil
	}
	afterSave := func(ctx context.Context, placement *entity.Placement) error {
		status := map[string]string{
			"draft": "placement_process", "pending_verification": "placement_process",
			"approved": "awaiting_documents", "ready": "ready", "active": "active",
			"completed": "completed", "cancelled": "unplaced", "transferred": "transferred",
		}[placement.Status]
		return db.WithContext(ctx).Table("students").Where("id = ?", placement.StudentID).Update("pkl_status", status).Error
	}
	service := crud.NewService("placement", repo, auditor, validate, nil).WithAfterSave(afterSave)
	handler := crud.NewHandler(service, "period_id", "student_id", "company_id", "supervisor_id", "status", "work_system")
	group := api.Group("/placements")
	crud.RegisterRoutes(group, handler, require, "placement")
	transferHandler := placementhttp.NewHandler(placementservice.NewTransferService(db))
	group.POST("/:id/transfer", require("placement.transfer"), transferHandler.Transfer)
}

func invalidPlacement(code, message string) error {
	return &apperrors.AppError{Status: http.StatusUnprocessableEntity, Code: code, Message: message}
}
