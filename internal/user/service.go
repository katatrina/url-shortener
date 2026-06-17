package user

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

const bcryptCost = 12

type Service struct {
	userRepo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{userRepo: repo}
}

func (s *Service) Signup(ctx context.Context, params SignupParams) (*User, error) {
	id, _ := uuid.NewV7()

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(params.Password), bcryptCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
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
