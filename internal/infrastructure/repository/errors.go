package repository

import (
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
)

var (
	ErrForeignKeyViolated = errors.New("foreign key constraint violated")
)

func TranslateGormError(err error) error {
	if err == nil {
		return nil
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23505":
			return gorm.ErrDuplicatedKey
		case "23503":
			return ErrForeignKeyViolated
		}
	}

	return err
}
