package model

import "errors"

var (
	ErrURLNotFound      = errors.New("url not found")
	ErrURLExpired       = errors.New("url has expired")
	ErrShortCodeTaken   = errors.New("short code is already taken")
	ErrInvalidShortCode = errors.New("short code contains invalid characters")
	ErrInvalidURL       = errors.New("invalid url format")
	ErrURLOwnerMismatch = errors.New("url does not belong to user")

	ErrUserNotFound         = errors.New("user not found")
	ErrEmailAlreadyExists   = errors.New("email already exists")
	ErrIncorrectCredentials = errors.New("incorrect email or password")
)
