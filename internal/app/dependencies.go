package app

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	"simpkl-api/internal/config"
	"simpkl-api/internal/platform/database"
	platformlogger "simpkl-api/internal/platform/logger"
)

type Dependencies struct {
	Config   *config.Config
	Logger   *zap.Logger
	Database *database.Connection
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

	return &Dependencies{
		Config:   cfg,
		Logger:   log,
		Database: connection,
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
