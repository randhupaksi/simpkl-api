package crud

import (
	"context"
	"net/http"

	"simpkl-api/internal/shared/errors"
	"simpkl-api/internal/shared/pagination"
	"simpkl-api/internal/shared/types"
)

type ValidationFunc[T any] func(context.Context, *T, *T) error
type NormalizeFunc[T any] func(*T)
type AfterSaveFunc[T any] func(context.Context, *T) error

type Service[T any] struct {
	resource  string
	repo      Repository[T]
	auditor   types.Auditor
	validate  ValidationFunc[T]
	normalize NormalizeFunc[T]
	afterSave AfterSaveFunc[T]
}

func NewService[T any](
	resource string,
	repo Repository[T],
	auditor types.Auditor,
	validate ValidationFunc[T],
	normalize NormalizeFunc[T],
) *Service[T] {
	if auditor == nil {
		auditor = types.NoopAuditor{}
	}
	return &Service[T]{
		resource: resource, repo: repo, auditor: auditor,
		validate: validate, normalize: normalize,
	}
}

func (s *Service[T]) WithAfterSave(afterSave AfterSaveFunc[T]) *Service[T] {
	s.afterSave = afterSave
	return s
}

func (s *Service[T]) List(
	ctx context.Context,
	query pagination.Query,
	filters map[string]string,
) ([]T, pagination.Meta, error) {
	query.Normalize()
	items, total, err := s.repo.List(ctx, query, filters)
	return items, pagination.NewMeta(query, total), mapError(err)
}

func (s *Service[T]) Get(ctx context.Context, id string) (*T, error) {
	item, err := s.repo.Get(ctx, id)
	return item, mapError(err)
}

func (s *Service[T]) Create(
	ctx context.Context,
	input *T,
	event types.AuditEvent,
) (*T, error) {
	if s.normalize != nil {
		s.normalize(input)
	}
	if s.validate != nil {
		if err := s.validate(ctx, nil, input); err != nil {
			return nil, err
		}
	}
	if err := s.repo.Create(ctx, input); err != nil {
		return nil, mapError(err)
	}
	if s.afterSave != nil {
		if err := s.afterSave(ctx, input); err != nil {
			return nil, err
		}
	}
	event.Action = "create"
	event.Resource = s.resource
	event.After = input
	_ = s.auditor.Record(event)
	return input, nil
}

func (s *Service[T]) Update(
	ctx context.Context,
	id string,
	input *T,
	event types.AuditEvent,
) (*T, error) {
	existing, err := s.repo.Get(ctx, id)
	if err != nil {
		return nil, mapError(err)
	}
	if s.normalize != nil {
		s.normalize(input)
	}
	if s.validate != nil {
		if err := s.validate(ctx, existing, input); err != nil {
			return nil, err
		}
	}
	updated, err := s.repo.Update(ctx, id, input)
	if err != nil {
		return nil, mapError(err)
	}
	if s.afterSave != nil {
		if err := s.afterSave(ctx, updated); err != nil {
			return nil, err
		}
	}
	event.Action = "update"
	event.Resource = s.resource
	event.ResourceID = id
	event.Before = existing
	event.After = updated
	_ = s.auditor.Record(event)
	return updated, nil
}

func (s *Service[T]) Delete(
	ctx context.Context,
	id string,
	event types.AuditEvent,
) error {
	existing, err := s.repo.Get(ctx, id)
	if err != nil {
		return mapError(err)
	}
	if err := s.repo.Delete(ctx, id); err != nil {
		return mapError(err)
	}
	event.Action = "delete"
	event.Resource = s.resource
	event.ResourceID = id
	event.Before = existing
	_ = s.auditor.Record(event)
	return nil
}

func mapError(err error) error {
	if err == nil {
		return nil
	}
	if IsNotFound(err) {
		return &errors.AppError{
			Status:  http.StatusNotFound,
			Code:    "RESOURCE_NOT_FOUND",
			Message: "Data tidak ditemukan",
			Cause:   err,
		}
	}
	return err
}
