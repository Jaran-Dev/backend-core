package db

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Config struct {
	DSN             string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

func NewPostgres(ctx context.Context, cfg Config) (*gorm.DB, error) {
	gormCfg := &gorm.Config{
		PrepareStmt: true,
		Logger: logger.New(
			slog.NewLogLogger(slog.Default().Handler(), slog.LevelInfo),
			logger.Config{
				SlowThreshold:             200 * time.Millisecond,
				LogLevel:                  logger.Warn,
				IgnoreRecordNotFoundError: true, // We handle ErrRecordNotFound in our mappers
				Colorful:                  false,
			},
		),
	}

	db, err := gorm.Open(postgres.Open(cfg.DSN), gormCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(cfg.ConnMaxLifetime)

	if err := sqlDB.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping db: %w", err)
	}

	slog.Info("Successfully connected to PostgreSQL via GORM")
	return db, nil
}
