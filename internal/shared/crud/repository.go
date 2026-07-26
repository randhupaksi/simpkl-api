package crud

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"gorm.io/gorm"

	"simpkl-api/internal/shared/pagination"
)

type Repository[T any] interface {
	List(context.Context, pagination.Query, map[string]string) ([]T, int64, error)
	Get(context.Context, string) (*T, error)
	Create(context.Context, *T) error
	Update(context.Context, string, *T) (*T, error)
	Delete(context.Context, string) error
}

type GormRepository[T any] struct {
	db            *gorm.DB
	searchColumns []string
	filterColumns map[string]string
	preloads      []string
}

func NewGormRepository[T any](
	db *gorm.DB,
	searchColumns []string,
	filterColumns map[string]string,
	preloads ...string,
) *GormRepository[T] {
	return &GormRepository[T]{
		db:            db,
		searchColumns: searchColumns,
		filterColumns: filterColumns,
		preloads:      preloads,
	}
}

func (r *GormRepository[T]) List(
	ctx context.Context,
	query pagination.Query,
	filters map[string]string,
) ([]T, int64, error) {
	query.Normalize()
	statement := r.db.WithContext(ctx).Model(new(T))

	if query.Search != "" && len(r.searchColumns) > 0 {
		conditions := make([]string, len(r.searchColumns))
		values := make([]any, len(r.searchColumns))
		for index, column := range r.searchColumns {
			conditions[index] = fmt.Sprintf("%s LIKE ?", column)
			values[index] = "%" + query.Search + "%"
		}
		statement = statement.Where(strings.Join(conditions, " OR "), values...)
	}

	for key, value := range filters {
		column, allowed := r.filterColumns[key]
		if allowed && value != "" {
			statement = statement.Where(column+" = ?", value)
		}
	}

	var total int64
	if err := statement.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var result []T
	statement = r.withPreloads(statement)
	if err := statement.
		Order("created_at DESC").
		Offset(query.Offset()).
		Limit(query.PerPage).
		Find(&result).Error; err != nil {
		return nil, 0, err
	}
	return result, total, nil
}

func (r *GormRepository[T]) Get(ctx context.Context, id string) (*T, error) {
	var result T
	statement := r.withPreloads(r.db.WithContext(ctx))
	if err := statement.First(&result, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &result, nil
}

func (r *GormRepository[T]) Create(ctx context.Context, entity *T) error {
	return r.db.WithContext(ctx).Create(entity).Error
}

func (r *GormRepository[T]) Update(ctx context.Context, id string, input *T) (*T, error) {
	existing, err := r.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	if err := r.db.WithContext(ctx).
		Model(existing).
		Select("*").
		Omit("id", "created_at", "deleted_at").
		Updates(input).Error; err != nil {
		return nil, err
	}
	return r.Get(ctx, id)
}

func (r *GormRepository[T]) Delete(ctx context.Context, id string) error {
	result := r.db.WithContext(ctx).Delete(new(T), "id = ?", id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *GormRepository[T]) withPreloads(statement *gorm.DB) *gorm.DB {
	for _, preload := range r.preloads {
		statement = statement.Preload(preload)
	}
	return statement
}

func IsNotFound(err error) bool {
	return errors.Is(err, gorm.ErrRecordNotFound)
}
