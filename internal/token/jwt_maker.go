package token

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

var (
	ErrTokenExpired = errors.New("token has expired")
	ErrTokenInvalid = errors.New("invalid token")
)

type JWTMaker struct {
	secretKey string
	ttl       time.Duration
}

func NewJWTMaker(secretKey string, ttl time.Duration) *JWTMaker {
	return &JWTMaker{secretKey, ttl}
}

func (m *JWTMaker) CreateToken(userID string) (string, time.Time, error) {
	now := time.Now()
	expiresAt := time.Now().Add(m.ttl)
	claims := jwt.RegisteredClaims{
		Subject:   userID,
		ExpiresAt: jwt.NewNumericDate(expiresAt),
		NotBefore: jwt.NewNumericDate(now),
		IssuedAt:  jwt.NewNumericDate(now),
		ID:        uuid.NewString(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString([]byte(m.secretKey))
	if err != nil {
		return "", time.Time{}, err
	}

	return tokenStr, expiresAt, nil
}

func (m *JWTMaker) VerifyToken(tokenStr string) (string, error) {
	t, err := jwt.ParseWithClaims(
		tokenStr,
		&jwt.RegisteredClaims{},
		func(_ *jwt.Token) (any, error) {
			return []byte(m.secretKey), nil
		},
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
	)
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return "", ErrTokenExpired
		}
		return "", ErrTokenInvalid
	}

	claims, ok := t.Claims.(*jwt.RegisteredClaims)
	if !ok {
		return "", ErrTokenInvalid
	}

	return claims.Subject, nil
}
