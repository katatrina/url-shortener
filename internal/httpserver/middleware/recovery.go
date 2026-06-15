package middleware

import (
	"log/slog"
	"net/http"
	"runtime/debug"

	"github.com/gin-gonic/gin"
)

// Recovery .
func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			r := recover()
			if r == nil {
				return
			}

			ctx := c.Request.Context()

			slog.Error("panic recovered",
				"request_id", RequestIDFromContext(ctx),
				"method", c.Request.Method,
				"path", c.Request.URL.Path,
				"panic", r,
				"stack", string(debug.Stack()),
			)

			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"error": "internal server error",
			})
		}()

		c.Next()
	}
}
