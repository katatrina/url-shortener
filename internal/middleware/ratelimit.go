package middleware

import (
	"fmt"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis_rate/v10"
	"github.com/katatrina/url-shortener/internal/logger"
	"github.com/katatrina/url-shortener/internal/response"
)

func RateLimit(limiter *redis_rate.Limiter, limit redis_rate.Limit) gin.HandlerFunc {
	return func(c *gin.Context) {
		log := logger.FromRequestContext(c.Request.Context())

		// Lib redis_rate always adds a fixed prefix "rate" to every key.
		key := fmt.Sprintf("%s:%s", "shorten", c.ClientIP())

		// Allow runs a Lua script on Redis that atomically checks the limit,
		// increments the counter, and sets TTL matching the window period.
		result, err := limiter.Allow(c.Request.Context(), key, limit)
		if err != nil {
			// Redis down — let the request through (Fail open).
			// Else, Fail closed = block.
			log.Warn("rate limit check failed", "error", err)
			c.Next()
			return
		}

		// Set custom rate limit headers so clients can self-throttle.
		c.Header("RateLimit-Limit", strconv.Itoa(limit.Rate))
		c.Header("RateLimit-Remaining", strconv.Itoa(max(0, result.Remaining)))
		c.Header("RateLimit-Reset", strconv.Itoa(int(result.ResetAfter.Seconds())))

		if result.Allowed == 0 {
			c.Header("Retry-After", strconv.Itoa(int(result.RetryAfter.Seconds())))
			response.TooManyRequests(c, "Rate limit exceeded. Try again later")
			c.Abort()
			return
		}

		c.Next()
	}
}
