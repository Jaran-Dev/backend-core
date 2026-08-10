package subscription

import (
	"context"
	"errors"
	"fmt"
	"time"

	domainSubscription "github.com/Jaran-Dev/backend-core/internal/domain/subscription"
	domainUser "github.com/Jaran-Dev/backend-core/internal/domain/user"
	"github.com/Jaran-Dev/backend-core/internal/infrastructure/repository"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type pgRepository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) domainSubscription.Repository {
	return &pgRepository{pool: pool}
}

func (r *pgRepository) Create(ctx context.Context, s *domainSubscription.Subscription) error {
	query := `
        INSERT INTO subscriptions (id, user_id, source, starts_at, ends_at, provider_ref, created_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7)
    `
	_, err := r.pool.Exec(ctx, query,
		uuid.UUID(s.ID()),
		uuid.UUID(s.UserID()),
		string(s.Source()),
		s.StartsAt(),
		repository.NullableTime(s.EndsAt()),
		repository.NullableString(s.ProviderRef()),
		s.CreatedAt(),
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return domainSubscription.ErrAlreadySubscribed
		}
		return fmt.Errorf("failed to create subscription: %w", err)
	}
	return nil
}

func (r *pgRepository) GetActiveByUserID(ctx context.Context, userID domainUser.UserID) (*domainSubscription.Subscription, error) {
	query := `
        SELECT id, user_id, source, starts_at, ends_at, provider_ref, created_at
        FROM subscriptions
        WHERE user_id = $1 
          AND starts_at <= now() 
          AND (ends_at IS NULL OR ends_at > now())
        ORDER BY created_at DESC 
        LIMIT 1
    `
	row := r.pool.QueryRow(ctx, query, uuid.UUID(userID))

	var (
		id          uuid.UUID
		uid         uuid.UUID
		sourceStr   string
		startsAt    time.Time
		endsAt      *time.Time
		providerRef *string
		createdAt   time.Time
	)

	err := row.Scan(&id, &uid, &sourceStr, &startsAt, &endsAt, &providerRef, &createdAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domainSubscription.ErrSubscriptionNotFound
		}
		return nil, fmt.Errorf("failed to get active subscription: %w", err)
	}

	return domainSubscription.UnmarshalSubscription(id, domainUser.UserID(uid), sourceStr, startsAt, endsAt, providerRef, createdAt)
}

func (r *pgRepository) ListByUserID(ctx context.Context, userID domainUser.UserID) ([]*domainSubscription.Subscription, error) {
	query := `
        SELECT id, user_id, source, starts_at, ends_at, provider_ref, created_at
        FROM subscriptions
        WHERE user_id = $1
        ORDER BY created_at DESC
    `
	rows, err := r.pool.Query(ctx, query, uuid.UUID(userID))
	if err != nil {
		return nil, fmt.Errorf("failed to list subscriptions: %w", err)
	}
	defer rows.Close()

	var subs []*domainSubscription.Subscription
	for rows.Next() {
		var (
			id          uuid.UUID
			uid         uuid.UUID
			sourceStr   string
			startsAt    time.Time
			endsAt      *time.Time
			providerRef *string
			createdAt   time.Time
		)
		if err := rows.Scan(&id, &uid, &sourceStr, &startsAt, &endsAt, &providerRef, &createdAt); err != nil {
			return nil, err
		}
		s, err := domainSubscription.UnmarshalSubscription(id, domainUser.UserID(uid), sourceStr, startsAt, endsAt, providerRef, createdAt)
		if err != nil {
			return nil, err
		}
		subs = append(subs, s)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating subscription rows: %w", err)
	}

	return subs, nil
}
