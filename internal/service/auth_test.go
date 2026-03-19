package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/katatrina/url-shortener/internal/mock"
	"github.com/katatrina/url-shortener/internal/model"
	"go.uber.org/mock/gomock"
	"golang.org/x/crypto/bcrypt"
)

// --- Register ---

func TestRegister_Success(t *testing.T) {
	ctrl := gomock.NewController(t)

	mockUserRepo := mock.NewMockUserRepository(ctrl)
	svc := New(nil, mockUserRepo, nil, nil)

	// We can't predict the exact user object (ID is random, password is hashed),
	// so we use gomock.Any() and capture the argument to verify it.
	mockUserRepo.EXPECT().
		Create(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, user model.User) (*model.User, error) {
			// Verify the password was hashed, not stored as plain text.
			err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte("securepass123"))
			if err != nil {
				t.Error("password was not properly hashed")
			}

			if user.Email != "alice@example.com" {
				t.Errorf("expected email alice@example.com, got %s", user.Email)
			}

			if user.DisplayName != "Alice" {
				t.Errorf("expected display name Alice, got %s", user.DisplayName)
			}

			// Simulate DB returning the created user.
			return &user, nil
		})

	user, err := svc.Register(context.Background(), model.CreateUserParams{
		Email:       "alice@example.com",
		DisplayName: "Alice",
		Password:    "securepass123",
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if user.Email != "alice@example.com" {
		t.Errorf("expected email alice@example.com, got %s", user.Email)
	}
}

func TestRegister_EmailAlreadyExists(t *testing.T) {
	ctrl := gomock.NewController(t)

	mockUserRepo := mock.NewMockUserRepository(ctrl)
	svc := New(nil, mockUserRepo, nil, nil)

	mockUserRepo.EXPECT().
		Create(gomock.Any(), gomock.Any()).
		Return(nil, model.ErrEmailAlreadyExists)

	_, err := svc.Register(context.Background(), model.CreateUserParams{
		Email:       "taken@example.com",
		DisplayName: "Bob",
		Password:    "securepass123",
	})

	if !errors.Is(err, model.ErrEmailAlreadyExists) {
		t.Errorf("expected ErrEmailAlreadyExists, got %v", err)
	}
}

// --- Login ---

func TestLogin_Success(t *testing.T) {
	ctrl := gomock.NewController(t)

	mockUserRepo := mock.NewMockUserRepository(ctrl)
	mockTokenMaker := mock.NewMockTokenMaker(ctrl)
	svc := New(nil, mockUserRepo, nil, mockTokenMaker)

	// Pre-hash the password as it would be stored in DB.
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("correctpass"), bcrypt.DefaultCost)

	storedUser := &model.User{
		ID:           "user-123",
		Email:        "alice@example.com",
		DisplayName:  "Alice",
		PasswordHash: string(hashedPassword),
	}

	expiresAt := time.Now().Add(24 * time.Hour)

	mockUserRepo.EXPECT().
		FindByEmail(gomock.Any(), "alice@example.com").
		Return(storedUser, nil)

	mockTokenMaker.EXPECT().
		CreateToken("user-123").
		Return("jwt-token-string", expiresAt, nil)

	result, err := svc.Login(context.Background(), model.LoginUserParams{
		Email:    "alice@example.com",
		Password: "correctpass",
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result.AccessToken != "jwt-token-string" {
		t.Errorf("expected token jwt-token-string, got %s", result.AccessToken)
	}
	if result.User.ID != "user-123" {
		t.Errorf("expected user ID user-123, got %s", result.User.ID)
	}
}

func TestLogin_UserNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)

	mockUserRepo := mock.NewMockUserRepository(ctrl)
	svc := New(nil, mockUserRepo, nil, nil)

	mockUserRepo.EXPECT().
		FindByEmail(gomock.Any(), "ghost@example.com").
		Return(nil, model.ErrUserNotFound)

	// Token maker should NOT be called if user doesn't exist.

	_, err := svc.Login(context.Background(), model.LoginUserParams{
		Email:    "ghost@example.com",
		Password: "whatever",
	})

	// Should return ErrIncorrectCredentials, NOT ErrUserNotFound.
	// This is a security practice: don't reveal whether the email exists.
	if !errors.Is(err, model.ErrIncorrectCredentials) {
		t.Errorf("expected ErrIncorrectCredentials, got %v", err)
	}
}

func TestLogin_WrongPassword(t *testing.T) {
	ctrl := gomock.NewController(t)

	mockUserRepo := mock.NewMockUserRepository(ctrl)
	svc := New(nil, mockUserRepo, nil, nil)

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("correctpass"), bcrypt.DefaultCost)

	storedUser := &model.User{
		ID:           "user-123",
		Email:        "alice@example.com",
		PasswordHash: string(hashedPassword),
	}

	mockUserRepo.EXPECT().
		FindByEmail(gomock.Any(), "alice@example.com").
		Return(storedUser, nil)

	_, err := svc.Login(context.Background(), model.LoginUserParams{
		Email:    "alice@example.com",
		Password: "wrongpass",
	})

	if !errors.Is(err, model.ErrIncorrectCredentials) {
		t.Errorf("expected ErrIncorrectCredentials, got %v", err)
	}
}
