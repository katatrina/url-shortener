package token

import (
	"errors"
	"time"
)

var (
	ErrTokenExpired = errors.New("token has expired")
	ErrTokenInvalid = errors.New("token is invalid")
)

// TokenMaker defines the interface for creating and verifying tokens.
type TokenMaker interface {
	CreateToken(userID string) (string, time.Time, error)
	VerifyToken(tokenString string) (string, error)
}
