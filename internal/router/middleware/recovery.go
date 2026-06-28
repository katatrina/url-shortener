package middleware

import (
	"log/slog"
	"runtime/debug"

	"github.com/gin-gonic/gin"
	"github.com/katatrina/url-shortener/internal/response"
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

			slog.ErrorContext(ctx, "panic recovered",
				"method", c.Request.Method,
				"path", c.Request.URL.Path,
				"panic", r,
				"stack", string(debug.Stack()),
			)

			response.Internal(c)
		}()

		c.Next()
	}
}
