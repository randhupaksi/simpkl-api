package logger

import (
	"fmt"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"simpkl-api/internal/config"
)

func New(cfg config.LogConfig, environment config.Environment) (*zap.Logger, error) {
	level := zapcore.DebugLevel
	if err := level.UnmarshalText([]byte(cfg.Level)); err != nil {
		return nil, fmt.Errorf("parse log level %q: %w", cfg.Level, err)
	}

	zapConfig := zap.NewDevelopmentConfig()
	if environment.IsProduction() {
		zapConfig = zap.NewProductionConfig()
	}
	zapConfig.Level = zap.NewAtomicLevelAt(level)

	log, err := zapConfig.Build()
	if err != nil {
		return nil, fmt.Errorf("build logger: %w", err)
	}

	return log, nil
}
