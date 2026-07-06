package user

import "errors"

var (
	ErrEmailExists          = errors.New("email already exists")
	ErrUserNotFound         = errors.New("user not found")
	ErrCredentialsIncorrect = errors.New("incorrect email or password")
)
