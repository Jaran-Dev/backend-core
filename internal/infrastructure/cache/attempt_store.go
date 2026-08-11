package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/Jaran-Dev/backend-core/internal/domain/account"
	"github.com/redis/go-redis/v9"
)

type RedisAttemptStore struct {
	client *redis.Client
	window time.Duration
}

func NewRedisAttemptStore(client *redis.Client, window time.Duration) *RedisAttemptStore {
	return &RedisAttemptStore{client: client, window: window}
}

func (s *RedisAttemptStore) IncrementFailedAttempt(ctx context.Context, email string) (int, error) {
	key := fmt.Sprintf("login_attempts:%s", email)
	count, err := s.client.Incr(ctx, key).Result()
	if err != nil {
		return 0, err
	}
	if count == 1 && s.window > 0 {
		s.client.Expire(ctx, key, s.window)
	}
	return int(count), nil
}

func (s *RedisAttemptStore) ResetFailedAttempts(ctx context.Context, email string) error {
	key := fmt.Sprintf("login_attempts:%s", email)
	return s.client.Del(ctx, key).Err()
}

func (s *RedisAttemptStore) LockAccount(ctx context.Context, accID account.AccountID, duration time.Duration) error {
	key := fmt.Sprintf("lockout:%s", accID.String())
	return s.client.Set(ctx, key, "locked", duration).Err()
}

func (s *RedisAttemptStore) IsAccountLocked(ctx context.Context, accID account.AccountID) (bool, error) {
	key := fmt.Sprintf("lockout:%s", accID.String())
	val, err := s.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return val == "locked", nil
}

func (s *RedisAttemptStore) UnlockAccount(ctx context.Context, accID account.AccountID) error {
	key := fmt.Sprintf("lockout:%s", accID.String())
	return s.client.Del(ctx, key).Err()
}
