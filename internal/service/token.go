package service

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

const (
	MinSecretKeySize = 32
	MinTokenExpiry   = 5 * time.Minute
)

var (
	ErrTokenExpired = errors.New("token has expired")
	ErrTokenInvalid = errors.New("token is invalid")
)

// JWTMaker implements TokenMaker using HS256.
type JWTMaker struct {
	secretKey []byte
	expiry    time.Duration
}

func NewJWTMaker(secretKey []byte, expiry time.Duration) (*JWTMaker, error) {
	if len(secretKey) < MinSecretKeySize {
		return nil, fmt.Errorf("secret key must be at least %d bytes", MinSecretKeySize)
	}

	if expiry < MinTokenExpiry {
		return nil, fmt.Errorf("token expiry must be at least %s", MinTokenExpiry)
	}

	return &JWTMaker{
		secretKey: secretKey,
		expiry:    expiry,
	}, nil
}

func (m *JWTMaker) CreateToken(userID string) (string, time.Time, error) {
	now := time.Now()
	expiresAt := now.Add(m.expiry)

	claims := jwt.RegisteredClaims{
		Subject:   userID,
		ExpiresAt: jwt.NewNumericDate(expiresAt),
		IssuedAt:  jwt.NewNumericDate(now),
		NotBefore: jwt.NewNumericDate(now),
		ID:        uuid.NewString(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString(m.secretKey)
	if err != nil {
		return "", time.Time{}, err
	}

	return tokenStr, expiresAt, nil
}

func (m *JWTMaker) VerifyToken(tokenString string) (*TokenClaims, error) {
	token, err := jwt.ParseWithClaims(
		tokenString,
		&jwt.RegisteredClaims{},
		func(token *jwt.Token) (any, error) {
			return m.secretKey, nil
		},
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
	)

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrTokenExpired
		}
		return nil, ErrTokenInvalid
	}

	claims, ok := token.Claims.(*jwt.RegisteredClaims)
	if !ok {
		return nil, ErrTokenInvalid
	}

	return &TokenClaims{
		UserID:    claims.Subject,
		ExpiresAt: claims.ExpiresAt.Time,
	}, nil
}
