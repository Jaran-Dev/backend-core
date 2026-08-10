package user

import (
	"context"
	"errors"
	"fmt"
	"time"

	domainUser "github.com/Jaran-Dev/backend-core/internal/domain/user"
	"github.com/Jaran-Dev/backend-core/internal/infrastructure/repository"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type pgRepository struct {
	db repository.DBTX
}

func NewRepository(db repository.DBTX) domainUser.UserRepository {
	return &pgRepository{db: db}
}

func (r *pgRepository) Create(ctx context.Context, u *domainUser.User) error {
	query := `
        INSERT INTO users (id, email, password_hash, display_name, ui_locale, created_at)
        VALUES ($1, $2, $3, $4, $5, $6)
    `
	_, err := r.db.Exec(ctx, query,
		uuid.UUID(u.ID()),
		u.Email().String(),
		u.PasswordHash(),
		repository.NullableString(u.DisplayName()),
		u.UILocale(),
		u.CreatedAt(),
	)
	if err != nil {
		// Check for Postgres unique violation (code 23505)
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return domainUser.ErrUserAlreadyExists
		}
		return fmt.Errorf("failed to create user: %w", err)
	}
	return nil
}

func (r *pgRepository) GetByID(ctx context.Context, id domainUser.UserID) (*domainUser.User, error) {
	query := `
        SELECT id, email, password_hash, display_name, ui_locale, created_at
        FROM users
        WHERE id = $1
    `
	row := r.db.QueryRow(ctx, query, uuid.UUID(id))
	return r.scanUser(row)
}

func (r *pgRepository) GetByEmail(ctx context.Context, email domainUser.Email) (*domainUser.User, error) {
	query := `
        SELECT id, email, password_hash, display_name, ui_locale, created_at
        FROM users
        WHERE email = $1
    `
	row := r.db.QueryRow(ctx, query, email.String())
	return r.scanUser(row)
}

type rowScanner interface {
	Scan(dest ...interface{}) error
}

func (r *pgRepository) scanUser(row rowScanner) (*domainUser.User, error) {
	var (
		id           uuid.UUID
		emailStr     string
		passwordHash string
		displayName  *string // Use pointer to handle SQL NULL cleanly
		uiLocale     string
		createdAt    time.Time
	)

	err := row.Scan(
		&id,
		&emailStr,
		&passwordHash,
		&displayName,
		&uiLocale,
		&createdAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domainUser.ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to scan user row: %w", err)
	}

	return domainUser.UnmarshalUser(id, emailStr, passwordHash, displayName, uiLocale, createdAt)
}
