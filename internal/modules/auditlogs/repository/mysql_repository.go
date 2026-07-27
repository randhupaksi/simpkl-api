package repository

import (
	"context"

	"gorm.io/gorm"

	"simpkl-api/internal/modules/auditlogs/entity"
	"simpkl-api/internal/shared/crud"
	"simpkl-api/internal/shared/pagination"
)

type MySQLRepository struct {
	base *crud.GormRepository[entity.AuditLog]
	db   *gorm.DB
}

func NewMySQLRepository(db *gorm.DB) *MySQLRepository {
	return &MySQLRepository{
		db: db,
		base: crud.NewGormRepository[entity.AuditLog](
			db,
			[]string{"action", "resource", "resource_id", "request_id"},
			map[string]string{
				"actor_id":   "actor_id",
				"action":     "action",
				"resource":   "resource",
				"request_id": "request_id",
			},
		),
	}
}

func (r *MySQLRepository) Create(ctx context.Context, log *entity.AuditLog) error {
	return r.db.WithContext(ctx).Create(log).Error
}

func (r *MySQLRepository) List(ctx context.Context, query pagination.Query, filters map[string]string) ([]entity.AuditLog, int64, error) {
	return r.base.List(ctx, query, filters)
}

func (r *MySQLRepository) Get(ctx context.Context, id string) (*entity.AuditLog, error) {
	return r.base.Get(ctx, id)
}
