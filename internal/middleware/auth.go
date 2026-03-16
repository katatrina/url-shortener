package middleware

import (
	"errors"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/katatrina/url-shortener/internal/response"
	"github.com/katatrina/url-shortener/internal/token"
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

func Auth(tokenMaker token.TokenMaker) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Unauthorized(c, response.CodeAuthRequired,
				"Authorization header is required")
			c.Abort()
			return
		}

		scheme, tokenString, found := strings.Cut(authHeader, " ")
		if !found || scheme != "Bearer" || tokenString == "" {
			response.Unauthorized(c, response.CodeAuthRequired,
				"Authorization header format must be: Bearer {token}")
			c.Abort()
			return
		}

		userID, err := tokenMaker.VerifyToken(tokenString)
		if err != nil {
			switch {
			case errors.Is(err, token.ErrTokenExpired):
				response.Unauthorized(c, response.CodeTokenExpired,
					"Token has expired")
			default:
				response.Unauthorized(c, response.CodeTokenInvalid,
					"Invalid token")
			}
			c.Abort()
			return
		}

		c.Set(AuthUserIDKey, userID)
		c.Next()
	}
}
