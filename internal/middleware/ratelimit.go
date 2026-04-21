package middleware

import (
	"fmt"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis_rate/v10"
	"github.com/katatrina/url-shortener/internal/logger"
	"github.com/katatrina/url-shortener/internal/response"
)

// RateLimit giới hạn request theo IP, dùng token bucket trên Redis.
//
// keyPrefix giúp phân biệt các rate limiter khác nhau trong app
// (e.g. "shorten", "login"), tránh share counter giữa các route không liên quan.
//
// Khi Redis down → fail-open (cho qua). Rationale: thà user abuse vài phút
// còn hơn toàn bộ service sập vì rate limiter hỏng. Quyết định này phù hợp
// với service URL shortener (low-risk). Service high-risk (payment, auth)
// nên fail-closed.
func RateLimit(limiter *redis_rate.Limiter, keyPrefix string, limit redis_rate.Limit) gin.HandlerFunc {
	return func(c *gin.Context) {
		log := logger.FromRequestContext(c.Request.Context())

		// redis_rate tự prefix "rate:" vào mọi key. Key cuối cùng sẽ là
		// "rate:shorten:<ip>". Không cần lo collision với key khác trong Redis.
		key := fmt.Sprintf("%s:%s", keyPrefix, c.ClientIP())

		result, err := limiter.Allow(c.Request.Context(), key, limit)
		if err != nil {
			log.Warn("rate limit check failed, failing open",
				"error", err, "key_prefix", keyPrefix, "ip", c.ClientIP())
			c.Next()
			return
		}

		// Expose header theo convention RFC draft-polli-ratelimit-headers.
		// Reset = số giây cho đến khi cửa sổ reset (không phải epoch timestamp).
		c.Header("RateLimit-Limit", strconv.Itoa(limit.Rate))
		c.Header("RateLimit-Remaining", strconv.Itoa(max(0, result.Remaining)))
		c.Header("RateLimit-Reset", strconv.Itoa(int(result.ResetAfter.Seconds())))

		if result.Allowed == 0 {
			c.Header("Retry-After", strconv.Itoa(int(result.RetryAfter.Seconds())))
			log.Info("rate limit exceeded", "key_prefix", keyPrefix, "ip", c.ClientIP())
			response.TooManyRequests(c, "Rate limit exceeded. Try again later")
			c.Abort()
			return
		}

		c.Next()
	}
}
