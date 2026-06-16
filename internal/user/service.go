package user

import (
	"context"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type Service struct {
	userRepo *Repository
}

func (s *Service) Signup(ctx context.Context, params SignupParams) (*User, error) {
	id, _ := uuid.NewV7()

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(params.Password), 12)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	user, err := s.userRepo.Create(ctx, User{
		ID:           id.String(),
		Email:        params.Email,
		PasswordHash: string(passwordHash),
		UpdatedAt:    now,
		CreatedAt:    now,
	})
	if err != nil {
		return nil, err
	}

	return user, nil
}
