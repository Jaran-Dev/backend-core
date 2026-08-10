package repository

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type DBTX interface {
	Exec(ctx context.Context, sql string, arguments ...interface{}) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row
}

func NullableString(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}

func DerefString(s *string) string {
	if s != nil {
		return *s
	}
	return ""
}

func NullableTime(t time.Time) interface{} {
	if t.IsZero() {
		return nil
	}
	return t
}
