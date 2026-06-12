package middleware

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// RequestID .
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		id, _ := uuid.NewV7()
		requestID := id.String()

		ctx := context.WithValue(c.Request.Context(), requestIDKey, requestID)
		c.Request = c.Request.WithContext(ctx)

		c.Header("X-Request-ID", requestID)

		c.Next()
	}
}

// RequestIDFromContext .
func RequestIDFromContext(ctx context.Context) string {
	id, _ := ctx.Value(requestIDKey).(string)
	return id
}
