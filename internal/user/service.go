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

type SignupParams struct {
	Email    string
	Password string
}

type AuthResult struct {
	AccessToken string
	User        *User
}

func (s *Service) newAuthResult(user *User) (*AuthResult, error) {
	accessToken, err := s.tokenIssuer.Issue(user.ID)
	if err != nil {
		return nil, fmt.Errorf("issue token: %w", err)
	}

	return &AuthResult{
		AccessToken: accessToken,
		User:        user,
	}, nil
}

func (s *Service) Signup(ctx context.Context, arg SignupParams) (*AuthResult, error) {
	id, _ := uuid.NewV7()

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(arg.Password), bcryptCost)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	user, err := s.userRepo.Insert(ctx, InsertUserCommand{
		ID:           id.String(),
		Email:        arg.Email,
		PasswordHash: string(passwordHash),
	})
	if err != nil {
		return nil, fmt.Errorf("insert user: %w", err)
	}

	return s.newAuthResult(user)
}

type LoginParams struct {
	Email    string
	Password string
}

func (s *Service) Login(ctx context.Context, arg LoginParams) (*AuthResult, error) {
	user, err := s.userRepo.FindByEmail(ctx, arg.Email)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return nil, ErrCredentialsIncorrect
		}
		return nil, fmt.Errorf("find user by email: %w", err)
	}

	if user.PasswordHash == nil {
		return nil, errors.New("user has no password hash")
	}

	if err = bcrypt.CompareHashAndPassword([]byte(*user.PasswordHash), []byte(arg.Password)); err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return nil, ErrCredentialsIncorrect
		}
		return nil, fmt.Errorf("compare password: %w", err)
	}

	return s.newAuthResult(user)
}

func (s *Service) GetUser(ctx context.Context, id string) (*User, error) {
	user, err := s.userRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("find user by id: %w", err)
	}

	return user, nil
}
