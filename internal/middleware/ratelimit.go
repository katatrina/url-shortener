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
		key := fmt.Sprintf("rl:%s", c.ClientIP())

		result, err := limiter.Allow(c.Request.Context(), key, limit)
		if err != nil {
			// Redis down — let the request through.
			log.Printf("[WARN] rate limit check failed: %v", err)
			c.Next()
			return
		}

		// Set standard rate limit headers.
		c.Header("X-RateLimit-Limit", strconv.Itoa(limit.Rate))
		c.Header("X-RateLimit-Remaining", strconv.Itoa(max(0, result.Remaining)))
		c.Header("Retry-After", strconv.Itoa(int(result.RetryAfter.Seconds())))

		if result.Allowed == 0 {
			response.TooManyRequests(c, "Rate limit exceeded. Try again later")
			c.Abort()
			return
		}

		c.Next()
	}
}
