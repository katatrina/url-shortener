package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func CORS(allowedOrigins ...string) gin.HandlerFunc {
	allowed := make(map[string]struct{}, len(allowedOrigins))
	for _, o := range allowedOrigins {
		allowed[o] = struct{}{}
	}

	return func(c *gin.Context) {
		c.Header("Vary", "Origin")

		origin := c.GetHeader("Origin")
		_, originAllowed := allowed[origin]
		if originAllowed {
			c.Header("Access-Control-Allow-Origin", origin)
		}

		// A CORS preflight is precisely an OPTIONS request carrying
		// Access-Control-Request-Method.
		isPreflight := c.Request.Method == http.MethodOptions &&
			c.GetHeader("Access-Control-Request-Method") != ""

		if isPreflight {
			// Preflight-only response headers; pointless on actual responses.
			if originAllowed {
				c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
				c.Header("Access-Control-Allow-Headers", "Authorization, Content-Type")
				c.Header("Access-Control-Max-Age", "86400")
			}

			// Answer the preflight regardless of origin: a disallowed origin
			// gets 204 without Allow-Origin, which the browser blocks.
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
