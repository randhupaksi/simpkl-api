package companies

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	companyhttp "simpkl-api/internal/modules/companies/delivery/http"
	"simpkl-api/internal/modules/companies/entity"
	companyservice "simpkl-api/internal/modules/companies/service"
	"simpkl-api/internal/shared/crud"
	apperrors "simpkl-api/internal/shared/errors"
	"simpkl-api/internal/shared/types"
)

func Register(api *gin.RouterGroup, db *gorm.DB, auditor types.Auditor, require func(string) gin.HandlerFunc) {
	repo := crud.NewGormRepository[entity.Company](
		db,
		[]string{"name", "industry", "city", "phone", "email"},
		map[string]string{"status": "status", "industry": "industry", "city": "city"},
	)
	validate := func(_ context.Context, _ *entity.Company, company *entity.Company) error {
		if company.CooperationStart != nil && company.CooperationEnd != nil &&
			company.CooperationEnd.Before(*company.CooperationStart) {
			return &apperrors.AppError{Status: http.StatusUnprocessableEntity, Code: "INVALID_DATE_RANGE", Message: "Tanggal akhir kerja sama harus setelah tanggal mulai"}
		}
		return nil
	}
	service := crud.NewService("company", repo, auditor, validate, nil)
	handler := crud.NewHandler(service, "status", "industry", "city")
	group := api.Group("/companies")
	partnerships := companyhttp.NewHandler(companyservice.NewPartnershipService(db))
	group.GET("/eligible", require("company.view"), partnerships.EligibleCompanies)
	group.GET("/:id/major-capacities", require("company.view"), partnerships.MajorCapacities)
	group.PUT("/:id/major-capacities", require("company.update"), partnerships.SetMajorCapacities)
	crud.RegisterRoutes(group, handler, require, "company")
}
