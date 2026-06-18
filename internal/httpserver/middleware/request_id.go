package middleware

import (
	"log/slog"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/katatrina/url-shortener/internal/logger"
)

// RequestID .
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		id, _ := uuid.NewV7()
		requestID := id.String()

		ctx := c.Request.Context()
		ctx = logger.WithAttrs(ctx, slog.String("request_id", requestID))

		c.Request = c.Request.WithContext(ctx)
		c.Header("X-Request-ID", requestID)

		c.Next()
	}
}
