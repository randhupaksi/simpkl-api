package repository

import (
	"context"

	authentity "simpkl-api/internal/modules/auth/entity"
	userentity "simpkl-api/internal/modules/users/entity"
)

type Repository interface {
	FindByLogin(context.Context, string) (*userentity.User, error)
	FindByID(context.Context, string) (*userentity.User, error)
	LoadAccess(context.Context, string) ([]string, []string, error)
	LoadScope(context.Context, string) ([]string, string, string, error)
	HasPermission(context.Context, string, string) (bool, error)
	SaveRefreshSession(context.Context, *authentity.RefreshSession) error
	FindRefreshSession(context.Context, string) (*authentity.RefreshSession, error)
	RevokeRefreshSession(context.Context, string) error
	UpdateLastLogin(context.Context, string) error
}
