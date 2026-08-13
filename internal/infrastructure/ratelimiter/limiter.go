package ratelimiter

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type Config struct {
	Capacity   int
	RefillRate int
	Window     time.Duration
}

type Result struct {
	Allowed    bool
	Remaining  int
	RetryAfter time.Duration
}

type Limiter struct {
	client *redis.Client
	script *redis.Script
}

func NewLimiter(client *redis.Client) *Limiter {
	// This Lua script implements the Token Bucket algorithm atomically.
	// KEYS[1] = the rate limit key (e.g., "rl:ip:192.168.1.1:login")
	// ARGV[1] = capacity (max burst)
	// ARGV[2] = refill_rate (tokens per second)
	// ARGV[3] = now (current time in milliseconds)
	// ARGV[4] = cost (usually 1)
	// ARGV[5] = ttl (key expiration in seconds)
	script := redis.NewScript(`
        local key = KEYS[1]
        local capacity = tonumber(ARGV[1])
        local refill_rate = tonumber(ARGV[2])
        local now = tonumber(ARGV[3])
        local cost = tonumber(ARGV[4])
        local ttl = tonumber(ARGV[5])

        -- Get current bucket state
        local bucket = redis.call("HMGET", key, "tokens", "last_refill")
        local tokens = tonumber(bucket[1])
        local last_refill = tonumber(bucket[2])

        -- Initialize bucket if it doesn't exist
        if tokens == nil then
            tokens = capacity
            last_refill = now
        end

        -- Calculate tokens to add since last refill
        local delta_ms = math.max(0, now - last_refill)
        local delta_sec = delta_ms / 1000.0
        local refilled = math.floor(delta_sec * refill_rate)
        
        if refilled > 0 then
            tokens = math.min(capacity, tokens + refilled)
            last_refill = last_refill + (refilled * 1000) -- update last_refill accurately
        end

        local allowed = 0
        local retry_after = 0

        -- Check if we have enough tokens
        if tokens >= cost then
            tokens = tokens - cost
            allowed = 1
        else
            -- Calculate how long to wait for 'cost' tokens
            local tokens_needed = cost - tokens
            retry_after = math.ceil((tokens_needed / refill_rate) * 1000)
        end

        -- Save state back to Redis
        redis.call("HMSET", key, "tokens", tokens, "last_refill", last_refill)
        redis.call("EXPIRE", key, ttl)

        return {allowed, tokens, retry_after}
    `)

	return &Limiter{
		client: client,
		script: script,
	}
}

func (l *Limiter) Allow(ctx context.Context, key string, cfg Config) (*Result, error) {
	now := time.Now().UnixMilli()
	cost := 1
	ttl := int(cfg.Window.Seconds())

	result, err := l.script.Run(ctx, l.client, []string{key},
		cfg.Capacity, cfg.RefillRate, now, cost, ttl,
	).Result()

	if err != nil {
		return nil, fmt.Errorf("rate limiter script error: %w", err)
	}

	vals := result.([]interface{})
	allowed := vals[0].(int64) == 1
	remaining := int(vals[1].(int64))
	retryAfterMs := int(vals[2].(int64))

	return &Result{
		Allowed:    allowed,
		Remaining:  remaining,
		RetryAfter: time.Duration(retryAfterMs) * time.Millisecond,
	}, nil
}
