package middleware

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/katatrina/url-shortener/internal/logger"
)

// Logger ghi 1 dòng log mỗi request, mức độ tùy status code.
//
// Quy ước level:
//   5xx → ERROR (server lỗi, cần để mắt)
//   4xx → WARN  (client lỗi, theo dõi nhưng không báo động)
//   2xx, 3xx → INFO
//
// Middleware này phải đặt SAU RequestID để log có request_id.
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		log := logger.FromRequestContext(c.Request.Context())

		attrs := []any{
			slog.String("method", c.Request.Method),
			slog.String("path", c.Request.URL.Path),
			slog.String("query", c.Request.URL.RawQuery),
			slog.String("client_ip", c.ClientIP()),
			slog.Int("status", c.Writer.Status()),
			slog.Int("bytes", c.Writer.Size()),
			slog.Duration("duration", time.Since(start)),
		}

		// Gin lưu error qua c.Error() — gom hết vào log.
		// Handler của bạn không xài c.Error() nhưng Recovery middleware có.
		if len(c.Errors) > 0 {
			attrs = append(attrs, slog.String("errors", c.Errors.String()))
		}

		switch {
		case c.Writer.Status() >= 500:
			log.Error("request completed", attrs...)
		case c.Writer.Status() >= 400:
			log.Warn("request completed", attrs...)
		default:
			log.Info("request completed", attrs...)
		}
	}
}
