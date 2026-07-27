package companycontacts

import (
	"context"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"simpkl-api/internal/modules/companycontacts/entity"
	"simpkl-api/internal/shared/crud"
	"simpkl-api/internal/shared/types"
)

func Register(api *gin.RouterGroup, db *gorm.DB, auditor types.Auditor, require func(string) gin.HandlerFunc) {
	repo := crud.NewGormRepository[entity.CompanyContact](
		db,
		[]string{"name", "position", "division", "phone", "email"},
		map[string]string{"company_id": "company_id", "is_primary": "is_primary"},
	)
	validate := func(ctx context.Context, _ *entity.CompanyContact, contact *entity.CompanyContact) error {
		if contact.IsPrimary {
			return db.WithContext(ctx).
				Model(&entity.CompanyContact{}).
				Where("company_id = ? AND id <> ?", contact.CompanyID, contact.ID).
				Update("is_primary", false).Error
		}
		return nil
	}
	service := crud.NewService("company_contact", repo, auditor, validate, nil)
	handler := crud.NewHandler(service, "company_id", "is_primary")
	crud.RegisterRoutes(api.Group("/company-contacts"), handler, require, "company")
}
