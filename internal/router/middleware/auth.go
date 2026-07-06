package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/katatrina/url-shortener/internal/apperror"
	"github.com/katatrina/url-shortener/internal/response"
	"github.com/katatrina/url-shortener/internal/token"
)

const ctxKeyUserID = "userID"

func Auth(tokenIssuer *token.Issuer) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Fail(c, apperror.New(http.StatusUnauthorized, apperror.CodeUnauthorized,
				"Authorization header is required"))
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			response.Fail(c, apperror.New(http.StatusUnauthorized, apperror.CodeUnauthorized,
				"Authorization header format must be 'Bearer {token}'"))
			c.Abort()
			return
		}

		userID, err := tokenIssuer.Verify(parts[1])
		if err != nil {
			response.Fail(c, apperror.New(http.StatusUnauthorized, apperror.CodeUnauthorized,
				"Invalid or expired token"))
			c.Abort()
			return
		}

		c.Set(ctxKeyUserID, userID)
		c.Next()
	}
}

func UserID(c *gin.Context) string {
	userID, ok := c.Value(ctxKeyUserID).(string)
	if !ok {
		panic("UserID() called without Auth middleware in the chain")
	}
	return userID
}
