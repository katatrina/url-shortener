package middleware

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
)

// AccessLog .
func AccessLog() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		method := c.Request.Method
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		c.Next() // ──── toàn bộ phần còn lại của request chạy ở đây ────

		ctx := c.Request.Context()
		status := c.Writer.Status()

		attrs := []any{
			"request_id", RequestIDFromContext(ctx),
			"method", method,
			"path", path,
			"status", status,
			"duration_ms", float64(time.Since(start).Microseconds()) / 1000,
			"client_ip", c.ClientIP(),
		}
		if query != "" {
			attrs = append(attrs, "query", query)
		}
		if len(c.Errors) > 0 {
			attrs = append(attrs, "errors", c.Errors.String())
		}

		switch {
		case status >= 500:
			slog.Error("request completed", attrs...)
		case status >= 400:
			slog.Warn("request completed", attrs...)
		default:
			slog.Info("request completed", attrs...)
		}
	}
}
