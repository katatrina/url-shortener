package middleware

import (
	"regexp"

	"log/slog"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/katatrina/url-shortener/internal/logger"
)

// requestIDPattern: chỉ chấp nhận alphanumeric, dấu gạch ngang, dấu gạch dưới,
// dài 8-128 ký tự. Đủ cho UUID, ULID, snowflake, Datadog trace_id.
var requestIDPattern = regexp.MustCompile(`^[a-zA-Z0-9_-]{8,128}$`)

const requestIDHeader = "X-Request-ID"

// RequestID lấy hoặc sinh request ID và gắn vào context logger.
// Header X-Request-ID từ upstream (load balancer, API gateway) được tôn trọng
// nếu hợp lệ — giúp trace cross-service.
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		rid := c.GetHeader(requestIDHeader)
		if !requestIDPattern.MatchString(rid) {
			rid = uuid.NewString()
		}

		log := slog.Default().With(slog.String("request_id", rid))
		ctx := logger.WithContext(c.Request.Context(), log)
		c.Request = c.Request.WithContext(ctx)

		c.Header(requestIDHeader, rid)
		c.Next()
	}
}
