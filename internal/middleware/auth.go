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

// Auth is a strict authentication middleware.
// Requests without a valid token are rejected with 401.
func Auth(tokenMaker token.TokenMaker) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Unauthorized(c, response.CodeAuthRequired,
				"Authorization header is required")
			c.Abort()
			return
		}

		userID, err := extractAndVerifyToken(authHeader, tokenMaker)
		if err != nil {
			handleTokenError(c, err)
			c.Abort()
			return
		}

		c.Set(AuthUserIDKey, userID)
		c.Next()
	}
}

// OptionalAuth allows both authenticated and anonymous requests.
//   - No Authorization header → anonymous, proceed normally.
//   - Has header but invalid/expired → reject with 401.
//   - Has header and valid → set user ID, proceed.
func OptionalAuth(tokenMaker token.TokenMaker) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.Next() // Anonymous request, that's fine.
			return
		}

		userID, err := extractAndVerifyToken(authHeader, tokenMaker)
		if err != nil {
			handleTokenError(c, err)
			c.Abort()
			return
		}

		c.Set(AuthUserIDKey, userID)
		c.Next()
	}
}

// extractAndVerifyToken parses the Authorization header and verifies the token.
func extractAndVerifyToken(authHeader string, tokenMaker token.TokenMaker) (string, error) {
	scheme, tokenString, found := strings.Cut(authHeader, " ")
	if !found || scheme != "Bearer" || tokenString == "" {
		return "", errBadAuthFormat
	}

	return tokenMaker.VerifyToken(tokenString)
}

// handleTokenError sends the appropriate 401 response based on the error type.
func handleTokenError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, errBadAuthFormat):
		response.Unauthorized(c, response.CodeAuthRequired,
			"Authorization header format must be: Bearer {token}")
	case errors.Is(err, token.ErrTokenExpired):
		response.Unauthorized(c, response.CodeTokenExpired,
			"Token has expired")
	default:
		response.Unauthorized(c, response.CodeTokenInvalid,
			"Invalid token")
	}
}

var errBadAuthFormat = errors.New("invalid authorization header format")
