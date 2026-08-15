package cache

import (
	"context"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisStateStore struct {
	client *redis.Client
}

func NewRedisStateStore(client *redis.Client) *RedisStateStore {
	return &RedisStateStore{client: client}
}

func (s *RedisStateStore) Save(ctx context.Context, state, codeVerifier, redirectURL string) error {
	key := "oauth:state:" + state
	pipe := s.client.TxPipeline()
	pipe.HSet(ctx, key, "code_verifier", codeVerifier, "redirect_url", redirectURL)
	pipe.Expire(ctx, key, 10*time.Minute)
	_, err := pipe.Exec(ctx)
	return err
}

const consumeScript = `
local key = 'oauth:state:' .. KEYS[1]
local res = redis.call('HMGET', key, 'code_verifier', 'redirect_url')
if res[1] == false then return nil end
redis.call('DEL', key)
return res
`

func (s *RedisStateStore) Consume(ctx context.Context, state string) (string, string, error) {
	res, err := s.client.Eval(ctx, consumeScript, []string{state}).Result()
	if err != nil {
		if err == redis.Nil {
			return "", "", errors.New("invalid or expired oauth state")
		}
		return "", "", err
	}

	vals, ok := res.([]interface{})
	if !ok || len(vals) < 2 {
		return "", "", errors.New("invalid state data")
	}

	codeVerifier, _ := vals[0].(string)
	redirectURL, _ := vals[1].(string)

	if codeVerifier == "" {
		return "", "", errors.New("invalid or expired oauth state")
	}

	return codeVerifier, redirectURL, nil
}
