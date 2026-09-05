package http

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/Jaran-Dev/backend-core/internal/infrastructure/ratelimiter"
	"github.com/gin-gonic/gin"
)

type KeyExtractorFunc func(c *gin.Context) string

func IPKeyExtractor() KeyExtractorFunc {
	return func(c *gin.Context) string {
		// ! NOTE
		// In production behind a load balancer, ensure Gin is configured
		// with `router.SetTrustedProxies()` to trust X-Forwarded-For.
		return c.ClientIP()
	}
}

func AccountIDKeyExtractor() KeyExtractorFunc {
	return func(c *gin.Context) string {
		accID, err := GetAccountIDFromContext(c)
		if err == nil {
			return "acc:" + accID.String()
		}
		// Fallback to IP if account is not authenticated
		return "ip:" + c.ClientIP()
	}
}

func RateLimitMiddleware(limiter *ratelimiter.Limiter, cfg ratelimiter.Config, scope string, keyFn KeyExtractorFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		rawKey := keyFn(c)

		redisKey := "rl:" + rawKey + ":" + scope

		ctx, cancel := context.WithTimeout(c.Request.Context(), 100*time.Millisecond)
		defer cancel()

		res, err := limiter.Allow(ctx, redisKey, cfg)
		if err != nil {
			// ! NOTE
			// Fail-open strategy: If Redis is down, we don't block legitimate traffic.
			// (For highly sensitive endpoints like login, you might choose to fail-closed).
			c.Error(err)
			c.Next()
			return
		}

		c.Header("X-RateLimit-Limit", strconv.Itoa(cfg.Capacity))
		c.Header("X-RateLimit-Remaining", strconv.Itoa(res.Remaining))

		if !res.Allowed {
			retrySec := int(res.RetryAfter.Seconds())
			if retrySec <= 0 {
				retrySec = 1
			}
			c.Header("Retry-After", strconv.Itoa(retrySec))

			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error":       "rate limit exceeded",
				"retry_after": retrySec,
			})
			return
		}

		c.Next()
	}
}

// StrictCooldown is a helper for endpoints that should only be called once per window.
func StrictCooldown(window time.Duration) ratelimiter.Config {
	return ratelimiter.Config{
		Capacity:   1,
		RefillRate: 1,
		Window:     window,
	}
}

// BurstLimit is a helper for standard API limits.
func BurstLimit(capacity int, perMinute int) ratelimiter.Config {
	return ratelimiter.Config{
		Capacity:   capacity,
		RefillRate: perMinute / 60,
		Window:     time.Minute,
	}
}
