package middleware

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/katatrina/url-shortener/internal/logger"
)

// Logger .
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		log := logger.FromRequestContext(c.Request.Context())

		c.Next()

		duration := time.Since(start)
		status := c.Writer.Status()

		attrs := []any{
			slog.Int("status", status),
			slog.Duration("duration", duration),
			slog.String("method", c.Request.Method),
			slog.String("path", c.Request.URL.Path),
		}

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
