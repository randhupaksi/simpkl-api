package service

import (
	"context"
	"encoding/json"

	"simpkl-api/internal/modules/auditlogs/entity"
	"simpkl-api/internal/modules/auditlogs/repository"
	"simpkl-api/internal/shared/pagination"
	"simpkl-api/internal/shared/types"
)

type Service struct{ repository repository.Repository }

func New(repository repository.Repository) *Service { return &Service{repository} }

func (s *Service) Record(event types.AuditEvent) error {
	before, _ := json.Marshal(event.Before)
	after, _ := json.Marshal(event.After)
	return s.repository.Create(context.Background(), &entity.AuditLog{
		ActorID: event.ActorID, Action: event.Action, Resource: event.Resource,
		ResourceID: event.ResourceID, RequestID: event.RequestID,
		BeforeJSON: string(before), AfterJSON: string(after), Reason: event.Reason,
		IPAddress: event.IPAddress, UserAgent: event.UserAgent,
	})
}

func (s *Service) List(ctx context.Context, query pagination.Query, filters map[string]string) ([]entity.AuditLog, pagination.Meta, error) {
	query.Normalize()
	items, total, err := s.repository.List(ctx, query, filters)
	return items, pagination.NewMeta(query, total), err
}

func (s *Service) Get(ctx context.Context, id string) (*entity.AuditLog, error) {
	return s.repository.Get(ctx, id)
}
