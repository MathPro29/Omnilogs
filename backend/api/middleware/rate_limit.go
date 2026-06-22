package middleware

import (
	"strconv"
	"sync"
	"time"

	"intern-api/configs"
	"intern-api/utils"

	"github.com/gin-gonic/gin"
)

type windowCounter struct {
	Count   int
	ResetAt time.Time
}

type RateLimiter struct {
	enabled  bool
	limit    int
	window   time.Duration
	group    string
	mu       sync.Mutex
	counters map[string]windowCounter
}

func NewRateLimiter(enabled bool, limit int, window time.Duration, group string) *RateLimiter {
	if limit <= 0 {
		limit = 1
	}
	if window <= 0 {
		window = time.Minute
	}
	return &RateLimiter{
		enabled:  enabled,
		limit:    limit,
		window:   window,
		group:    group,
		counters: make(map[string]windowCounter),
	}
}

func (l *RateLimiter) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !l.enabled {
			c.Next()
			return
		}

		now := time.Now()
		key := c.ClientIP() + ":" + l.group

		l.mu.Lock()
		counter := l.counters[key]
		if counter.ResetAt.IsZero() || now.After(counter.ResetAt) {
			counter = windowCounter{Count: 0, ResetAt: now.Add(l.window)}
		}
		counter.Count++
		remaining := l.limit - counter.Count
		l.counters[key] = counter
		l.mu.Unlock()

		if remaining < 0 {
			remaining = 0
		}

		setRateLimitHeaders(c, l.limit, remaining, counter.ResetAt)
		if counter.Count > l.limit {
			utils.Error(c, 429, "RATE_LIMIT_EXCEEDED", "too many requests")
			c.Abort()
			return
		}

		c.Next()
	}
}

func setRateLimitHeaders(c *gin.Context, limit int, remaining int, resetAt time.Time) {
	c.Header("X-RateLimit-Limit", strconv.Itoa(limit))
	c.Header("X-RateLimit-Remaining", strconv.Itoa(remaining))
	c.Header("X-RateLimit-Reset", strconv.FormatInt(resetAt.Unix(), 10))
}

func GlobalRateLimit(env *configs.Env) gin.HandlerFunc {
	return NewRateLimiter(
		env.RateLimitEnabled,
		env.RateLimitRequests,
		time.Duration(env.RateLimitWindowSeconds)*time.Second,
		"global",
	).Middleware()
}

func AuthRateLimit(env *configs.Env) gin.HandlerFunc {
	return NewRateLimiter(
		env.RateLimitEnabled,
		env.AuthRateLimitRequests,
		time.Duration(env.AuthRateLimitWindowSeconds)*time.Second,
		"auth",
	).Middleware()
}

func UploadRateLimit(env *configs.Env) gin.HandlerFunc {
	return NewRateLimiter(
		env.RateLimitEnabled,
		env.UploadRateLimitRequests,
		time.Duration(env.UploadRateLimitWindowSeconds)*time.Second,
		"upload",
	).Middleware()
}
