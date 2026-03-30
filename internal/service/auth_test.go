package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/katatrina/url-shortener/internal/model"
	mock "github.com/katatrina/url-shortener/internal/service/mock"
	"go.uber.org/mock/gomock"
	"golang.org/x/crypto/bcrypt"
)

func hashPassword(t *testing.T, password string) string {
	t.Helper()
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}
	return string(hash)
}

func TestLogin_HappyPath(t *testing.T) {
	ctrl := gomock.NewController(t)
	userRepo := mock.NewMockUserRepository(ctrl)
	tokenMaker := mock.NewMocktokenCreator(ctrl)
	svc := New(nil, userRepo, nil, nil, nil, tokenMaker, nil)

	hashedPwd := hashPassword(t, "correct-password")
	expiresAt := time.Now().Add(24 * time.Hour)

	userRepo.EXPECT().
		FindByEmail(gomock.Any(), "user@example.com").
		Return(&model.User{
			ID:           "user-id-1",
			Email:        "user@example.com",
			DisplayName:  "Test User",
			PasswordHash: hashedPwd,
		}, nil)

	tokenMaker.EXPECT().
		CreateToken("user-id-1").
		Return("jwt-token-123", expiresAt, nil)

	result, err := svc.Login(context.Background(), model.LoginUserParams{
		Email:    "user@example.com",
		Password: "correct-password",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.AccessToken != "jwt-token-123" {
		t.Fatalf("got token %q, want %q", result.AccessToken, "jwt-token-123")
	}
	if result.User.ID != "user-id-1" {
		t.Fatalf("got user ID %q, want %q", result.User.ID, "user-id-1")
	}
}

func TestLogin_UserNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	userRepo := mock.NewMockUserRepository(ctrl)
	svc := New(nil, userRepo, nil, nil, nil, nil, nil)

	userRepo.EXPECT().
		FindByEmail(gomock.Any(), "nobody@example.com").
		Return(nil, model.ErrUserNotFound)

	_, err := svc.Login(context.Background(), model.LoginUserParams{
		Email:    "nobody@example.com",
		Password: "any-password",
	})
	if !errors.Is(err, model.ErrIncorrectCredentials) {
		t.Fatalf("got %v, want ErrIncorrectCredentials", err)
	}
}

func TestLogin_WrongPassword(t *testing.T) {
	ctrl := gomock.NewController(t)
	userRepo := mock.NewMockUserRepository(ctrl)
	svc := New(nil, userRepo, nil, nil, nil, nil, nil)

	hashedPwd := hashPassword(t, "correct-password")

	userRepo.EXPECT().
		FindByEmail(gomock.Any(), "user@example.com").
		Return(&model.User{
			ID:           "user-id-1",
			Email:        "user@example.com",
			PasswordHash: hashedPwd,
		}, nil)

	_, err := svc.Login(context.Background(), model.LoginUserParams{
		Email:    "user@example.com",
		Password: "wrong-password",
	})
	if !errors.Is(err, model.ErrIncorrectCredentials) {
		t.Fatalf("got %v, want ErrIncorrectCredentials", err)
	}
}

func TestLogin_DBError(t *testing.T) {
	ctrl := gomock.NewController(t)
	userRepo := mock.NewMockUserRepository(ctrl)
	svc := New(nil, userRepo, nil, nil, nil, nil, nil)

	dbErr := errors.New("connection refused")
	userRepo.EXPECT().
		FindByEmail(gomock.Any(), "user@example.com").
		Return(nil, dbErr)

	_, err := svc.Login(context.Background(), model.LoginUserParams{
		Email:    "user@example.com",
		Password: "any-password",
	})
	if !errors.Is(err, dbErr) {
		t.Fatalf("got %v, want db error", err)
	}
}

func TestLogin_TokenCreationError(t *testing.T) {
	ctrl := gomock.NewController(t)
	userRepo := mock.NewMockUserRepository(ctrl)
	tokenMaker := mock.NewMocktokenCreator(ctrl)
	svc := New(nil, userRepo, nil, nil, nil, tokenMaker, nil)

	hashedPwd := hashPassword(t, "correct-password")

	userRepo.EXPECT().
		FindByEmail(gomock.Any(), "user@example.com").
		Return(&model.User{
			ID:           "user-id-1",
			Email:        "user@example.com",
			PasswordHash: hashedPwd,
		}, nil)

	tokenMaker.EXPECT().
		CreateToken("user-id-1").
		Return("", time.Time{}, errors.New("signing key error"))

	_, err := svc.Login(context.Background(), model.LoginUserParams{
		Email:    "user@example.com",
		Password: "correct-password",
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
