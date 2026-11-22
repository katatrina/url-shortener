package user

import (
	"strings"
	"time"
)

type SignupRequest struct {
	Email    string `json:"email"    validate:"required,max=254,ascii_email"`
	Password string `json:"password" validate:"required,min=8,max=32,max_bytes=72"`
}

func (r *SignupRequest) Normalize() {
	r.Email = strings.ToLower(strings.TrimSpace(r.Email))
}

// No email format check on login: a bad email just fails as wrong credentials.
type LoginRequest struct {
	Email    string `json:"email"    validate:"required"`
	Password string `json:"password" validate:"required"`
}

func (r *LoginRequest) Normalize() {
	r.Email = strings.ToLower(strings.TrimSpace(r.Email))
}

type AuthResponse struct {
	AccessToken string       `json:"accessToken"`
	User        UserResponse `json:"user"`
}

type UserResponse struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	FullName  string    `json:"fullName"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

func newAuthResponse(r *AuthResult) AuthResponse {
	return AuthResponse{
		AccessToken: r.AccessToken,
		User:        newUserResponse(r.User),
	}
}

func newUserResponse(u *User) UserResponse {
	return UserResponse{
		ID:        u.ID,
		Email:     u.Email,
		FullName:  u.FullName,
		CreatedAt: u.CreatedAt.UTC(),
		UpdatedAt: u.UpdatedAt.UTC(),
	}
}
