package periods

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"simpkl-api/internal/modules/periods/entity"
	"simpkl-api/internal/shared/crud"
	apperrors "simpkl-api/internal/shared/errors"
	"simpkl-api/internal/shared/types"
)

var allowedTransitions = map[string]map[string]bool{
	"draft":       {"draft": true, "preparation": true},
	"preparation": {"preparation": true, "draft": true, "active": true},
	"active":      {"active": true, "completed": true},
	"completed":   {"completed": true, "archived": true},
	"archived":    {"archived": true},
}

func Register(api *gin.RouterGroup, db *gorm.DB, auditor types.Auditor, require func(string) gin.HandlerFunc) {
	repo := crud.NewGormRepository[entity.Period](
		db,
		[]string{"name", "academic_year", "semester"},
		map[string]string{"academic_year": "academic_year", "semester": "semester", "cohort": "cohort", "status": "status"},
	)
	validate := func(_ context.Context, existing *entity.Period, period *entity.Period) error {
		if !period.EndDate.After(period.StartDate) {
			return &apperrors.AppError{Status: http.StatusUnprocessableEntity, Code: "INVALID_DATE_RANGE", Message: "Tanggal selesai harus setelah tanggal mulai"}
		}
		if existing != nil && !allowedTransitions[existing.Status][period.Status] {
			return &apperrors.AppError{Status: http.StatusConflict, Code: "INVALID_STATUS_TRANSITION", Message: "Perubahan status periode tidak diizinkan"}
		}
		return nil
	}
	service := crud.NewService("period", repo, auditor, validate, nil)
	handler := crud.NewHandler(service, "academic_year", "semester", "cohort", "status")
	crud.RegisterRoutes(api.Group("/periods"), handler, require, "period")
}
