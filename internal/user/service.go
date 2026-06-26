package user

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/katatrina/url-shortener/internal/token"
	"golang.org/x/crypto/bcrypt"
)

const bcryptCost = 12

type Service struct {
	userRepo    *Repository
	tokenIssuer *token.Issuer
}

func NewService(userRepo *Repository, tokenIssuer *token.Issuer) *Service {
	return &Service{
		userRepo:    userRepo,
		tokenIssuer: tokenIssuer,
	}
}

func (s *Service) Signup(ctx context.Context, params SignupParams) (*User, error) {
	id, _ := uuid.NewV7()

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(params.Password), bcryptCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	user, err := s.userRepo.Create(ctx, User{
		ID:           id.String(),
		Email:        params.Email,
		PasswordHash: string(passwordHash),
	})
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (s *Service) Login(ctx context.Context, params LoginParams) (*LoginResult, error) {
	user, err := s.userRepo.FindByEmail(ctx, params.Email)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return nil, ErrCredentialsIncorrect
		}
		return nil, err
	}

	if err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(params.Password)); err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return nil, ErrCredentialsIncorrect
		}
		return nil, fmt.Errorf("failed to compare password: %w", err)
	}

	accessToken, expiresAt, err := s.tokenIssuer.Issue(user.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to create token: %w", err)
	}

	return &LoginResult{
		AccessToken:          accessToken,
		AccessTokenExpiresAt: expiresAt,
		User:                 user,
	}, nil
}
