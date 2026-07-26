package config

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfigValidate(t *testing.T) {
	cfg := &Config{
		App: AppConfig{
			Name: "SIMPKL API",
			Env:  EnvironmentTest,
			Port: 8080,
		},
		Database: DatabaseConfig{
			Host:                  "localhost",
			Port:                  3306,
			Name:                  "simpkl_test",
			User:                  "root",
			MaxOpenConnections:    10,
			MaxIdleConnections:    5,
			ConnectionMaxLifetime: time.Minute,
		},
		JWT: JWTConfig{
			AccessSecret:  "access-secret-for-test",
			RefreshSecret: "refresh-secret-for-test",
			AccessTTL:     15 * time.Minute,
			RefreshTTL:    24 * time.Hour,
		},
		CORS: CORSConfig{
			AllowedOrigins: []string{"http://localhost:5173"},
		},
	}

	require.NoError(t, cfg.Validate())

	cfg.JWT.RefreshSecret = cfg.JWT.AccessSecret
	err := cfg.Validate()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "must be different")
}
