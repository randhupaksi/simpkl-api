package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	mysqlconfig "github.com/go-sql-driver/mysql"
	gormmysql "gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"simpkl-api/internal/config"
)

type Connection struct {
	GORM *gorm.DB
	SQL  *sql.DB
}

func Open(ctx context.Context, cfg config.DatabaseConfig) (*Connection, error) {
	dsnConfig := mysqlconfig.Config{
		User:      cfg.User,
		Passwd:    cfg.Password,
		Net:       "tcp",
		Addr:      fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		DBName:    cfg.Name,
		ParseTime: true,
		Loc:       time.Local,
	}
	dsn := dsnConfig.FormatDSN()

	db, err := gorm.Open(gormmysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Error),
	})
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("get database handle: %w", err)
	}

	sqlDB.SetMaxOpenConns(cfg.MaxOpenConnections)
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConnections)
	sqlDB.SetConnMaxLifetime(cfg.ConnectionMaxLifetime)

	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := sqlDB.PingContext(pingCtx); err != nil {
		_ = sqlDB.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}

	return &Connection{GORM: db, SQL: sqlDB}, nil
}

func (c *Connection) Close() error {
	if c == nil || c.SQL == nil {
		return nil
	}

	return c.SQL.Close()
}
