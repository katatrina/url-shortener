package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/katatrina/url-shortener/internal/model"
	"golang.org/x/crypto/bcrypt"
)

func (s *Service) Register(ctx context.Context, params model.CreateUserParams) (*model.User, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(params.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	userID, err := uuid.NewV7()
	if err != nil {
		return nil, fmt.Errorf("failed to generate user ID: %w", err)
	}

	now := time.Now()
	user := model.User{
		ID:           userID.String(),
		Email:        params.Email,
		FullName:     params.FullName,
		PasswordHash: string(hashedPassword),
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	created, err := s.userRepo.Create(ctx, user)
	if err != nil {
		return nil, err
	}

	return created, nil
}

func (s *Service) Login(ctx context.Context, params model.LoginUserParams) (*model.LoginUserResult, error) {
	user, err := s.userRepo.FindByEmail(ctx, params.Email)
	if err != nil {
		if errors.Is(err, model.ErrUserNotFound) {
			return nil, model.ErrIncorrectCredentials
		}
		return nil, err
	}

	if err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(params.Password)); err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return nil, model.ErrIncorrectCredentials
		}
		return nil, fmt.Errorf("failed to compare password: %w", err)
	}

	accessToken, expiresAt, err := s.tokenMaker.CreateToken(user.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to create token: %w", err)
	}

	return &model.LoginUserResult{
		AccessToken:          accessToken,
		AccessTokenExpiresAt: expiresAt,
		User:                 user,
	}, nil
}
