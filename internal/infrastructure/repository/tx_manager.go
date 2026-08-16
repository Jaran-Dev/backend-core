package repository

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"math/rand"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
)

type txKey struct{}

type TransactionManager struct {
	db *gorm.DB
}

func NewTransactionManager(db *gorm.DB) *TransactionManager {
	return &TransactionManager{db: db}
}

func (tm *TransactionManager) Do(ctx context.Context, fn func(ctx context.Context) error) error {
	maxRetries := 3
	var lastErr error

	for i := 0; i < maxRetries; i++ {
		err := tm.db.Transaction(func(tx *gorm.DB) error {
			txCtx := context.WithValue(ctx, txKey{}, tx)
			return fn(txCtx)
		}, &sql.TxOptions{
			Isolation: sql.LevelSerializable,
			ReadOnly:  false,
		})

		if err == nil {
			return nil
		}
		
		lastErr = err

		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "40001" {
			backoff := time.Duration(10*(i+1)) * time.Millisecond
			jitter := time.Duration(rand.Intn(10)) * time.Millisecond
			time.Sleep(backoff + jitter)

			slog.Warn("Transaction serialization failure, retrying...", "attempt", i+1)
			continue
		}

		return err
	}

	return lastErr
}

func GetDB(ctx context.Context, defaultDB *gorm.DB) *gorm.DB {
	if tx, ok := ctx.Value(txKey{}).(*gorm.DB); ok {
		return tx.WithContext(ctx)
	}
	return defaultDB.WithContext(ctx)
}
