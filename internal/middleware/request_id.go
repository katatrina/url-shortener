package middleware

import (
	"log/slog"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/katatrina/url-shortener/internal/logger"
)

// RequestID creates an unique ID for each request and attaches into context logger.
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = uuid.NewString()
		}

		// Tạo child logger: mọi log.Info(), log.Error() sau đó
		// đều tự động kèm {"request_id": "..."}
		log := slog.Default().With(slog.String("request_id", requestID))

		ctx := logger.WithContext(c.Request.Context(), log)
		c.Request = c.Request.WithContext(ctx)

		// Response header — client có thể dùng ID này để report bug
		// "Anh ơi, em gặp lỗi, request ID là xyz" → bạn search Loki ngay
		c.Header("X-Request-ID", requestID)

		c.Next()
	}
}
