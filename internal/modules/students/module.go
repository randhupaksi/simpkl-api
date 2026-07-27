package students

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	studenthttp "simpkl-api/internal/modules/students/delivery/http"
	"simpkl-api/internal/modules/students/entity"
	studentservice "simpkl-api/internal/modules/students/service"
	"simpkl-api/internal/shared/crud"
	"simpkl-api/internal/shared/types"
)

func Register(api *gin.RouterGroup, db *gorm.DB, auditor types.Auditor, require func(string) gin.HandlerFunc) {
	repo := crud.NewGormRepository[entity.Student](
		db,
		[]string{"nis", "nisn", "name", "email", "phone"},
		map[string]string{
			"class_id": "class_id", "major_id": "major_id", "cohort": "cohort",
			"status": "status", "pkl_status": "pkl_status",
		},
	)
	service := crud.NewService("student", repo, auditor, nil, nil)
	handler := crud.NewHandler(service, "class_id", "major_id", "cohort", "status", "pkl_status")
	group := api.Group("/students")
	crud.RegisterRoutes(group, handler, require, "student")
	importer := studenthttp.NewHandler(studentservice.NewImportService(db))
	group.POST("/import", require("student.import"), importer.Import)
}
