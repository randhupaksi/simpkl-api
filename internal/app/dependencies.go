package app

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	"simpkl-api/internal/config"
	auditrepository "simpkl-api/internal/modules/auditlogs/repository"
	auditservice "simpkl-api/internal/modules/auditlogs/service"
	authrepository "simpkl-api/internal/modules/auth/repository"
	platformauth "simpkl-api/internal/platform/auth"
	"simpkl-api/internal/platform/database"
	platformlogger "simpkl-api/internal/platform/logger"
	"simpkl-api/internal/platform/storage"
	"simpkl-api/internal/shared/types"
)

type Dependencies struct {
	Config         *config.Config
	Logger         *zap.Logger
	Database       *database.Connection
	Tokens         *platformauth.TokenManager
	AuthRepository authrepository.Repository
	Auditor        types.Auditor
	Storage        storage.Storage
}

func BuildDependencies(ctx context.Context) (*Dependencies, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("load config: %w", err)
	}

	log, err := platformlogger.New(cfg.Log, cfg.App.Env)
	if err != nil {
		return nil, err
	}

	connection, err := database.Open(ctx, cfg.Database)
	if err != nil {
		_ = log.Sync()
		return nil, err
	}

	fileStorage, err := storage.NewLocal(cfg.Storage.Path)
	if err != nil {
		_ = connection.Close()
		_ = log.Sync()
		return nil, fmt.Errorf("initialize private storage: %w", err)
	}
	tokenManager := platformauth.NewTokenManager(
		cfg.JWT.AccessSecret,
		cfg.JWT.RefreshSecret,
		cfg.JWT.AccessTTL,
		cfg.JWT.RefreshTTL,
	)
	authRepo := authrepository.NewMySQLRepository(connection.GORM)
	auditor := auditservice.New(auditrepository.NewMySQLRepository(connection.GORM))

	return &Dependencies{
		Config: cfg, Logger: log, Database: connection,
		Tokens: tokenManager, AuthRepository: authRepo,
		Auditor: auditor, Storage: fileStorage,
	}, nil
}

func (d *Dependencies) Close() {
	if d.Database != nil {
		if err := d.Database.Close(); err != nil {
			d.Logger.Error("failed to close database", zap.Error(err))
		}
	}
	if d.Logger != nil {
		_ = d.Logger.Sync()
	}
}
