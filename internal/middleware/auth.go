package middleware

import (
	"errors"
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/katatrina/url-shortener/internal/logger"
	"github.com/katatrina/url-shortener/internal/response"
	"github.com/katatrina/url-shortener/internal/token"
)

const AuthUserIDKey = "authUserID"

type tokenVerifier interface {
	VerifyToken(tokenStr string) (string, error)
}

// GetAuthUserID trả về user ID nếu có, nil nếu request anonymous.
// Dùng cho OptionalAuth route.
func GetAuthUserID(c *gin.Context) *string {
	val, exists := c.Get(AuthUserIDKey)
	if !exists {
		return nil
	}
	id, ok := val.(string)
	if !ok {
		return nil
	}
	return &id
}

// MustGetAuthUserID trả về user ID, panic nếu không có.
// Chỉ dùng trong route có Auth middleware attached — panic là programmer error.
func MustGetAuthUserID(c *gin.Context) string {
	val, exists := c.Get(AuthUserIDKey)
	if !exists {
		panic("MustGetAuthUserID called without Auth middleware — programmer error")
	}
	id, ok := val.(string)
	if !ok {
		panic(fmt.Sprintf("MustGetAuthUserID: expected string, got %T", val))
	}
	return id
}

// Auth yêu cầu token hợp lệ. Không có token hoặc token sai → 401.
func Auth(tv tokenVerifier) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Unauthorized(c, response.CodeAuthRequired,
				"Authorization header is required")
			c.Abort()
			return
		}

		userID, err := extractAndVerifyToken(authHeader, tv)
		if err != nil {
			handleTokenError(c, err)
			c.Abort()
			return
		}

		c.Set(AuthUserIDKey, userID)
		c.Next()
	}
}

// OptionalAuth cho phép cả anonymous và authenticated request.
//   Không có header      → cho qua, user anonymous.
//   Có header nhưng sai  → 401 (tránh trường hợp client nghĩ đã đăng nhập mà thực ra không).
//   Có header và đúng    → set user ID, cho qua.
func OptionalAuth(tv tokenVerifier) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.Next()
			return
		}

		userID, err := extractAndVerifyToken(authHeader, tv)
		if err != nil {
			handleTokenError(c, err)
			c.Abort()
			return
		}

		c.Set(AuthUserIDKey, userID)
		c.Next()
	}
}

func extractAndVerifyToken(authHeader string, tv tokenVerifier) (string, error) {
	scheme, tokenString, found := strings.Cut(authHeader, " ")
	if !found || scheme != "Bearer" || tokenString == "" {
		return "", errBadAuthFormat
	}
	return tv.VerifyToken(tokenString)
}

func handleTokenError(c *gin.Context, err error) {
	log := logger.FromRequestContext(c.Request.Context())

	switch {
	case errors.Is(err, errBadAuthFormat):
		log.Warn("auth failed: bad format")
		response.Unauthorized(c, response.CodeAuthRequired,
			"Authorization header format must be: Bearer {token}")
	case errors.Is(err, token.ErrTokenExpired):
		log.Info("auth failed: token expired")
		response.Unauthorized(c, response.CodeTokenExpired, "Token has expired")
	default:
		log.Warn("auth failed: invalid token", "error", err)
		response.Unauthorized(c, response.CodeTokenInvalid, "Invalid token")
	}
}

var errBadAuthFormat = errors.New("invalid authorization header format")
