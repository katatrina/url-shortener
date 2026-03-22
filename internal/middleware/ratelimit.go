package middleware

import (
	"fmt"
	"log"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis_rate/v10"
	"github.com/katatrina/url-shortener/internal/response"
)

func RateLimit(limiter *redis_rate.Limiter, limit redis_rate.Limit) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Lib redis_rate always adds a fixed prefix "rate" to every key.
		key := fmt.Sprintf("%s:%s", "shorten", c.ClientIP())

		result, err := limiter.Allow(c.Request.Context(), key, limit) // Allow handles entire rate limiting logic
		if err != nil {
			// Redis down — let the request through (Fail open).
			// Else, Fail closed = block.
			log.Printf("[WARN] rate limit check failed: %v", err)
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
