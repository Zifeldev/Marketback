package middleware

import (
	"fmt"
	"net/http"
	"time"

	"github.com/Zifeldev/marketback/service/Market/internal/cache"
	"github.com/gin-gonic/gin"
)

func RateLimiter(redis *cache.RedisCache, limit int, window time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		if redis == nil {
			c.Next()
			return
		}

		clientID := c.ClientIP()
		if userID, exists := c.Get("user_id"); exists {
			clientID = fmt.Sprintf("user:%v", userID)
		}

		key := fmt.Sprintf("ratelimit:%s", clientID)
		ctx := c.Request.Context()

		count, err := redis.Increment(ctx, key)
		if err != nil {
			c.Next()
			return
		}

		if count == 1 {
			_ = redis.Expire(ctx, key, window)
		}

		if count > int64(limit) {
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
