package admin

import (
	"context"
	"errors"
	"fmt"
	"time"

	domainAdmin "github.com/Jaran-Dev/backend-core/internal/domain/admin"
	"github.com/Jaran-Dev/backend-core/internal/infrastructure/repository"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type pgRepository struct {
	db repository.DBTX
}

func NewRepository(db repository.DBTX) domainAdmin.AdminRepository {
	return &pgRepository{db: db}
}

func (r *pgRepository) Create(ctx context.Context, a *domainAdmin.Admin) error {
	query := `
        INSERT INTO admins (id, email, password_hash, created_at)
        VALUES ($1, $2, $3, $4)
    `
	_, err := r.db.Exec(ctx, query,
		uuid.UUID(a.ID()),
		a.Email().String(),
		a.PasswordHash(),
		a.CreatedAt(),
	)
	if err != nil {
		// Check for Postgres unique violation (code 23505)
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return domainAdmin.ErrAdminAlreadyExists
		}
		return fmt.Errorf("failed to create admin: %w", err)
	}
	return nil
}

func (r *pgRepository) GetByID(ctx context.Context, id domainAdmin.AdminID) (*domainAdmin.Admin, error) {
	query := `
        SELECT id, email, password_hash, created_at
        FROM admins
        WHERE id = $1
    `
	row := r.db.QueryRow(ctx, query, uuid.UUID(id))
	return r.scanAdmin(row)
}

func (r *pgRepository) GetByEmail(ctx context.Context, email domainAdmin.Email) (*domainAdmin.Admin, error) {
	query := `
        SELECT id, email, password_hash, created_at
        FROM admins
        WHERE email = $1
    `
	row := r.db.QueryRow(ctx, query, email.String())
	return r.scanAdmin(row)
}

type rowScanner interface {
	Scan(dest ...interface{}) error
}

func (r *pgRepository) scanAdmin(row rowScanner) (*domainAdmin.Admin, error) {
	var (
		id           uuid.UUID
		emailStr     string
		passwordHash string
		createdAt    time.Time
	)

	err := row.Scan(
		&id,
		&emailStr,
		&passwordHash,
		&createdAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domainAdmin.ErrAdminNotFound
		}
		return nil, fmt.Errorf("failed to scan admin row: %w", err)
	}

	return domainAdmin.UnmarshalAdmin(id, emailStr, passwordHash, createdAt)
}
