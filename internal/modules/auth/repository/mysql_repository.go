package repository

import (
	"context"
	"database/sql"
	"time"

	"gorm.io/gorm"

	authentity "simpkl-api/internal/modules/auth/entity"
	userentity "simpkl-api/internal/modules/users/entity"
)

type MySQLRepository struct{ db *gorm.DB }

func NewMySQLRepository(db *gorm.DB) *MySQLRepository {
	return &MySQLRepository{db}
}

func (r *MySQLRepository) FindByLogin(
	ctx context.Context,
	login string,
) (*userentity.User, error) {
	var user userentity.User
	err := r.db.WithContext(ctx).
		Table("users").
		Where("(email = ? OR username = ?) AND status = ? AND deleted_at IS NULL", login, login, "active").
		First(&user).Error
	return &user, err
}

func (r *MySQLRepository) FindByID(
	ctx context.Context,
	id string,
) (*userentity.User, error) {
	var user userentity.User
	err := r.db.WithContext(ctx).First(&user, "id = ?", id).Error
	return &user, err
}

func (r *MySQLRepository) LoadAccess(
	ctx context.Context,
	userID string,
) ([]string, []string, error) {
	var roles []string
	if err := r.db.WithContext(ctx).
		Table("roles").
		Select("roles.code").
		Joins("JOIN user_roles ON user_roles.role_id = roles.id").
		Where("user_roles.user_id = ? AND roles.status = ?", userID, "active").
		Pluck("roles.code", &roles).Error; err != nil {
		return nil, nil, err
	}

	var permissions []string
	if err := r.db.WithContext(ctx).
		Table("permissions").
		Distinct("permissions.code").
		Joins("JOIN role_permissions ON role_permissions.permission_id = permissions.id").
		Joins("JOIN user_roles ON user_roles.role_id = role_permissions.role_id").
		Joins("JOIN roles ON roles.id = user_roles.role_id").
		Where("user_roles.user_id = ? AND roles.status = ?", userID, "active").
		Pluck("permissions.code", &permissions).Error; err != nil {
		return nil, nil, err
	}
	return roles, permissions, nil
}

func (r *MySQLRepository) LoadScope(
	ctx context.Context,
	userID string,
) ([]string, string, string, error) {
	var user struct {
		MajorID sql.NullString
		ClassID sql.NullString
	}
	if err := r.db.WithContext(ctx).Table("users").
		Select("major_id, class_id").
		Where("id = ? AND status = ? AND deleted_at IS NULL", userID, "active").
		Take(&user).Error; err != nil {
		return nil, "", "", err
	}
	roles, _, err := r.LoadAccess(ctx, userID)
	return roles, user.MajorID.String, user.ClassID.String, err
}

func (r *MySQLRepository) HasPermission(
	ctx context.Context,
	userID string,
	code string,
) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Table("permissions").
		Joins("JOIN role_permissions ON role_permissions.permission_id = permissions.id").
		Joins("JOIN user_roles ON user_roles.role_id = role_permissions.role_id").
		Joins("JOIN roles ON roles.id = user_roles.role_id").
		Where(
			"user_roles.user_id = ? AND roles.status = ? AND permissions.code IN ?",
			userID,
			"active",
			[]string{code, "*"},
		).
		Count(&count).Error
	return count > 0, err
}

func (r *MySQLRepository) SaveRefreshSession(
	ctx context.Context,
	session *authentity.RefreshSession,
) error {
	return r.db.WithContext(ctx).Create(session).Error
}

func (r *MySQLRepository) FindRefreshSession(
	ctx context.Context,
	hash string,
) (*authentity.RefreshSession, error) {
	var session authentity.RefreshSession
	err := r.db.WithContext(ctx).
		Where("token_hash = ? AND revoked_at IS NULL AND expires_at > ?", hash, time.Now()).
		First(&session).Error
	return &session, err
}

func (r *MySQLRepository) RevokeRefreshSession(
	ctx context.Context,
	hash string,
) error {
	now := time.Now()
	return r.db.WithContext(ctx).
		Model(&authentity.RefreshSession{}).
		Where("token_hash = ? AND revoked_at IS NULL", hash).
		Update("revoked_at", now).Error
}

func (r *MySQLRepository) UpdateLastLogin(ctx context.Context, userID string) error {
	return r.db.WithContext(ctx).
		Model(&userentity.User{}).
		Where("id = ?", userID).
		Update("last_login_at", time.Now()).Error
}
