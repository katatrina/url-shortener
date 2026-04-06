package middleware

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/katatrina/url-shortener/internal/logger"
)

// Logging records each request with full context.
//
// Separated from Metrics middleware due to separation of concerns:
// - Metrics: feeds Prometheus (metrics, aggregated)
// - Logging: feeds Loki (details of each request, debuggable)
//
// Both measure duration but serve different purposes.
func Logging() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// Lấy logger từ context — đã có request_id nhờ RequestID middleware
		log := logger.FromContext(c.Request.Context())

		log.Info("request started",
			slog.String("method", c.Request.Method),
			slog.String("path", c.Request.URL.Path),
			slog.String("client_ip", c.ClientIP()),
		)

		c.Next()

		duration := time.Since(start)
		status := c.Writer.Status()

		attrs := []any{
			slog.Int("status", status),
			slog.Duration("duration", duration),
			slog.String("method", c.Request.Method),
			slog.String("path", c.Request.URL.Path),
		}

		// Log level theo status code — giống learn-logging
		// 5xx → ERROR (cần attention), 4xx → WARN (client's fault), 2xx/3xx → INFO
		switch {
		case status >= 500:
			log.Error("request completed", attrs...)
		case status >= 400:
			log.Warn("request completed", attrs...)
		default:
			log.Info("request completed", attrs...)
		}
	}
}
