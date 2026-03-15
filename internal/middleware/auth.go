package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/katatrina/url-shortener/internal/service"
)

const AuthUserIDKey = "authUserID"

func GetAuthUserID(c *gin.Context) *string {
	val, exists := c.Get(AuthUserIDKey)
	if !exists {
		return nil
	}
	id := val.(string)
	return &id
}

func MustGetAuthUserID(c *gin.Context) string {
	val, exists := c.Get(AuthUserIDKey)
	if !exists {
		panic("MustGetAuthUserID: auth middleware not attached")
	}
	return val.(string)
}

func Auth(tokenMaker service.TokenMaker) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "authorization header is required",
			})
			return
		}

		scheme, tokenStr, found := strings.Cut(authHeader, " ")
		if !found || scheme != "Bearer" || tokenStr == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "authorization header format must be: Bearer {token}",
			})
			return
		}

		claims, err := tokenMaker.VerifyToken(tokenStr)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "invalid or expired token",
			})
			return
		}

		c.Set(AuthUserIDKey, claims.UserID)
		c.Next()
	}
}
