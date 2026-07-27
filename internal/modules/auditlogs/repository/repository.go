package repository

import (
	"context"

	"simpkl-api/internal/modules/auditlogs/entity"
	"simpkl-api/internal/shared/pagination"
)

type Repository interface {
	Create(context.Context, *entity.AuditLog) error
	List(context.Context, pagination.Query, map[string]string) ([]entity.AuditLog, int64, error)
	Get(context.Context, string) (*entity.AuditLog, error)
}
