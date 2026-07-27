package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"strings"
	"time"

	"gorm.io/gorm"

	authentity "simpkl-api/internal/modules/auth/entity"
	authrepository "simpkl-api/internal/modules/auth/repository"
	platformauth "simpkl-api/internal/platform/auth"
	apperrors "simpkl-api/internal/shared/errors"
)

type Service struct {
	repository authrepository.Repository
	tokens     *platformauth.TokenManager
	accessTTL  time.Duration
	refreshTTL time.Duration
}

func New(
	repository authrepository.Repository,
	tokens *platformauth.TokenManager,
	accessTTL time.Duration,
	refreshTTL time.Duration,
) *Service {
	return &Service{repository, tokens, accessTTL, refreshTTL}
}

func (s *Service) Login(
	ctx context.Context,
	login, password, ipAddress, userAgent string,
) (*authentity.AuthResult, error) {
	user, err := s.repository.FindByLogin(ctx, strings.TrimSpace(login))
	if err != nil || platformauth.ComparePassword(user.PasswordHash, password) != nil {
		return nil, unauthorized("Email/username atau password salah", err)
	}
	result, err := s.issue(ctx, user.ID, user.Email, ipAddress, userAgent)
	if err != nil {
		return nil, err
	}
	_ = s.repository.UpdateLastLogin(ctx, user.ID)
	result.User = s.profile(ctx, user.ID, user.Name, user.Email, user.Username, user.Status, user.MajorID, user.ClassID)
	return result, nil
}

func (s *Service) Refresh(
	ctx context.Context,
	refreshToken, ipAddress, userAgent string,
) (*authentity.AuthResult, error) {
	claims, err := s.tokens.ParseRefresh(refreshToken)
	if err != nil {
		return nil, unauthorized("Refresh token tidak valid", err)
	}
	hash := tokenHash(refreshToken)
	session, err := s.repository.FindRefreshSession(ctx, hash)
	if err != nil || session.UserID != claims.UserID {
		return nil, unauthorized("Refresh token tidak aktif", err)
	}
	user, err := s.repository.FindByID(ctx, claims.UserID)
	if err != nil || user.Status != "active" {
		return nil, unauthorized("Akun tidak aktif", err)
	}
	if err := s.repository.RevokeRefreshSession(ctx, hash); err != nil {
		return nil, err
	}
	result, err := s.issue(ctx, user.ID, user.Email, ipAddress, userAgent)
	if err != nil {
		return nil, err
	}
	result.User = s.profile(ctx, user.ID, user.Name, user.Email, user.Username, user.Status, user.MajorID, user.ClassID)
	return result, nil
}

func (s *Service) Logout(ctx context.Context, refreshToken string) error {
	if strings.TrimSpace(refreshToken) == "" {
		return nil
	}
	return s.repository.RevokeRefreshSession(ctx, tokenHash(refreshToken))
}

func (s *Service) Me(ctx context.Context, userID string) (*authentity.Profile, error) {
	user, err := s.repository.FindByID(ctx, userID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, unauthorized("Pengguna tidak ditemukan", err)
		}
		return nil, err
	}
	profile := s.profile(ctx, user.ID, user.Name, user.Email, user.Username, user.Status, user.MajorID, user.ClassID)
	return &profile, nil
}

func (s *Service) issue(
	ctx context.Context,
	userID, email, ipAddress, userAgent string,
) (*authentity.AuthResult, error) {
	accessToken, refreshToken, err := s.tokens.Issue(userID, email)
	if err != nil {
		return nil, err
	}
	session := &authentity.RefreshSession{
		UserID:    userID,
		TokenHash: tokenHash(refreshToken),
		ExpiresAt: time.Now().Add(s.refreshTTL),
		IPAddress: ipAddress,
		UserAgent: userAgent,
	}
	if err := s.repository.SaveRefreshSession(ctx, session); err != nil {
		return nil, err
	}
	return &authentity.AuthResult{
		Tokens: authentity.TokenPair{
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
			TokenType:    "Bearer",
			ExpiresIn:    int64(s.accessTTL.Seconds()),
		},
	}, nil
}

func (s *Service) profile(
	ctx context.Context,
	id, name, email, username, status, majorID, classID string,
) authentity.Profile {
	roles, permissions, _ := s.repository.LoadAccess(ctx, id)
	return authentity.Profile{
		ID: id, Name: name, Email: email, Username: username, Status: status,
		MajorID: majorID, ClassID: classID, Roles: roles, Permissions: permissions,
	}
}

func tokenHash(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func unauthorized(message string, cause error) error {
	return &apperrors.AppError{
		Status: http.StatusUnauthorized, Code: "UNAUTHORIZED", Message: message, Cause: cause,
	}
}
