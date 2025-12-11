package middleware

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/Zifeldev/marketback/service/Market/internal/cache"
	"github.com/Zifeldev/marketback/service/Market/internal/logger"
	"github.com/gin-gonic/gin"
)

type inMemoryLimiter struct {
	mu       sync.RWMutex
	counters map[string]*rateLimitEntry
	limit    int
	window   time.Duration
}

type rateLimitEntry struct {
	count     int
	expiresAt time.Time
}

func newInMemoryLimiter(limit int, window time.Duration) *inMemoryLimiter {
	limiter := &inMemoryLimiter{
		counters: make(map[string]*rateLimitEntry),
		limit:    limit,
		window:   window,
	}
	go limiter.cleanup()
	return limiter
}

func (l *inMemoryLimiter) cleanup() {
	ticker := time.NewTicker(time.Minute)
	for range ticker.C {
		l.mu.Lock()
		now := time.Now()
		for key, entry := range l.counters {
			if now.After(entry.expiresAt) {
				delete(l.counters, key)
			}
		}
		l.mu.Unlock()
	}
}

func (l *inMemoryLimiter) increment(key string) (int, bool) {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	entry, exists := l.counters[key]

	if !exists || now.After(entry.expiresAt) {
		l.counters[key] = &rateLimitEntry{
			count:     1,
			expiresAt: now.Add(l.window),
		}
		return 1, true
	}

	entry.count++
	return entry.count, entry.count <= l.limit
}

func RateLimiter(redis *cache.RedisCache, limit int, window time.Duration) gin.HandlerFunc {
	memLimiter := newInMemoryLimiter(limit, window)

	return func(c *gin.Context) {
		clientID := c.ClientIP()
		if userID, exists := c.Get("user_id"); exists {
			clientID = fmt.Sprintf("user:%v", userID)
		}

		key := fmt.Sprintf("ratelimit:%s", clientID)
		ctx := c.Request.Context()

		var count int64
		var allowed bool

		if redis != nil {
			redisCount, err := redis.Increment(ctx, key)
			if err != nil {
				logger.GetLogger().WithField("err", err).Warn("Redis rate limit failed, using in-memory fallback")
				memCount, ok := memLimiter.increment(key)
				count = int64(memCount)
				allowed = ok
			} else {
				count = redisCount
				if count == 1 {
					_ = redis.Expire(ctx, key, window)
				}
				allowed = count <= int64(limit)
			}
		} else {
			memCount, ok := memLimiter.increment(key)
			count = int64(memCount)
			allowed = ok
		}

		if !allowed {
			logger.GetLogger().WithFields(map[string]interface{}{
				"client_id": clientID,
				"count":     count,
				"limit":     limit,
			}).Warn("Rate limit exceeded")

			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":       "rate limit exceeded",
				"retry_after": window.Seconds(),
			})
			c.Abort()
			return
		}

		c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", limit))
		c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", limit-int(count)))

		c.Next()
	}
}
