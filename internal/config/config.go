package config

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	App      AppConfig
	Database DatabaseConfig
	JWT      JWTConfig
	CORS     CORSConfig
	Storage  StorageConfig
	Log      LogConfig
}

type AppConfig struct {
	Name string
	Env  Environment
	Port int
	URL  string
}

type DatabaseConfig struct {
	Host                  string
	Port                  int
	Name                  string
	User                  string
	Password              string
	MaxOpenConnections    int
	MaxIdleConnections    int
	ConnectionMaxLifetime time.Duration
}

type JWTConfig struct {
	AccessSecret  string
	RefreshSecret string
	AccessTTL     time.Duration
	RefreshTTL    time.Duration
}

type CORSConfig struct {
	AllowedOrigins []string
}

type StorageConfig struct {
	Driver string
	Path   string
}

type LogConfig struct {
	Level string
}

func Load() (*Config, error) {
	v := viper.New()
	v.SetConfigName(".env")
	v.SetConfigType("env")
	v.AddConfigPath(".")
	v.AutomaticEnv()

	setDefaults(v)

	if err := v.ReadInConfig(); err != nil {
		var notFound viper.ConfigFileNotFoundError
		if !errors.As(err, &notFound) {
			return nil, fmt.Errorf("read environment file: %w", err)
		}
	}

	cfg := &Config{
		App: AppConfig{
			Name: v.GetString("APP_NAME"),
			Env:  Environment(v.GetString("APP_ENV")),
			Port: v.GetInt("APP_PORT"),
			URL:  v.GetString("APP_URL"),
		},
		Database: DatabaseConfig{
			Host:                  v.GetString("DB_HOST"),
			Port:                  v.GetInt("DB_PORT"),
			Name:                  v.GetString("DB_NAME"),
			User:                  v.GetString("DB_USER"),
			Password:              v.GetString("DB_PASSWORD"),
			MaxOpenConnections:    v.GetInt("DB_MAX_OPEN_CONNECTIONS"),
			MaxIdleConnections:    v.GetInt("DB_MAX_IDLE_CONNECTIONS"),
			ConnectionMaxLifetime: time.Duration(v.GetInt("DB_CONNECTION_MAX_LIFETIME")) * time.Second,
		},
		JWT: JWTConfig{
			AccessSecret:  v.GetString("JWT_ACCESS_SECRET"),
			RefreshSecret: v.GetString("JWT_REFRESH_SECRET"),
			AccessTTL:     time.Duration(v.GetInt("JWT_ACCESS_TTL_MINUTES")) * time.Minute,
			RefreshTTL:    time.Duration(v.GetInt("JWT_REFRESH_TTL_HOURS")) * time.Hour,
		},
		CORS: CORSConfig{
			AllowedOrigins: splitAndTrim(v.GetString("CORS_ALLOWED_ORIGINS")),
		},
		Storage: StorageConfig{
			Driver: v.GetString("STORAGE_DRIVER"),
			Path:   v.GetString("STORAGE_PATH"),
		},
		Log: LogConfig{
			Level: v.GetString("LOG_LEVEL"),
		},
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

func (c *Config) Validate() error {
	var problems []string

	if strings.TrimSpace(c.App.Name) == "" {
		problems = append(problems, "APP_NAME is required")
	}
	if !c.App.Env.IsValid() {
		problems = append(problems, "APP_ENV must be development, test, or production")
	}
	if c.App.Port < 1 || c.App.Port > 65535 {
		problems = append(problems, "APP_PORT must be between 1 and 65535")
	}
	if c.Database.Host == "" || c.Database.Name == "" || c.Database.User == "" {
		problems = append(problems, "DB_HOST, DB_NAME, and DB_USER are required")
	}
	if c.Database.Port < 1 || c.Database.Port > 65535 {
		problems = append(problems, "DB_PORT must be between 1 and 65535")
	}
	if c.Database.MaxOpenConnections < 1 {
		problems = append(problems, "DB_MAX_OPEN_CONNECTIONS must be positive")
	}
	if c.Database.MaxIdleConnections < 0 ||
		c.Database.MaxIdleConnections > c.Database.MaxOpenConnections {
		problems = append(problems, "DB_MAX_IDLE_CONNECTIONS must be between 0 and DB_MAX_OPEN_CONNECTIONS")
	}
	if len(c.JWT.AccessSecret) < 16 || len(c.JWT.RefreshSecret) < 16 {
		problems = append(problems, "JWT secrets must contain at least 16 characters")
	}
	if c.JWT.AccessSecret == c.JWT.RefreshSecret {
		problems = append(problems, "JWT access and refresh secrets must be different")
	}
	if len(c.CORS.AllowedOrigins) == 0 {
		problems = append(problems, "CORS_ALLOWED_ORIGINS must contain at least one origin")
	}

	if len(problems) > 0 {
		return fmt.Errorf("invalid configuration: %s", strings.Join(problems, "; "))
	}

	return nil
}

func setDefaults(v *viper.Viper) {
	v.SetDefault("APP_NAME", "SIMPKL API")
	v.SetDefault("APP_ENV", string(EnvironmentDevelopment))
	v.SetDefault("APP_PORT", 8080)
	v.SetDefault("APP_URL", "http://localhost:8080")
	v.SetDefault("DB_HOST", "localhost")
	v.SetDefault("DB_PORT", 3306)
	v.SetDefault("DB_NAME", "simpkl")
	v.SetDefault("DB_USER", "root")
	v.SetDefault("DB_MAX_OPEN_CONNECTIONS", 25)
	v.SetDefault("DB_MAX_IDLE_CONNECTIONS", 10)
	v.SetDefault("DB_CONNECTION_MAX_LIFETIME", 300)
	v.SetDefault("JWT_ACCESS_TTL_MINUTES", 15)
	v.SetDefault("JWT_REFRESH_TTL_HOURS", 168)
	v.SetDefault("CORS_ALLOWED_ORIGINS", "http://localhost:5173")
	v.SetDefault("STORAGE_DRIVER", "local")
	v.SetDefault("STORAGE_PATH", "./storage/private")
	v.SetDefault("LOG_LEVEL", "debug")
}

func splitAndTrim(value string) []string {
	parts := strings.Split(value, ",")
	result := make([]string, 0, len(parts))

	for _, part := range parts {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			result = append(result, trimmed)
		}
	}

	return result
}
