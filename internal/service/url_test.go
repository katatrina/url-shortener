package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/katatrina/url-shortener/internal/model"
	mock "github.com/katatrina/url-shortener/internal/service/mock"
	"go.uber.org/mock/gomock"
)

func TestShortenURL_HappyPath(t *testing.T) {
	ctrl := gomock.NewController(t)
	urlRepo := mock.NewMockURLRepository(ctrl)
	svc := New(urlRepo, nil, nil, nil, nil, nil, nil)

	// Random code generation: ShortCodeExists returns false on first try.
	urlRepo.EXPECT().
		ShortCodeExists(gomock.Any(), gomock.Any()).
		Return(false, nil)

	urlRepo.EXPECT().
		Create(gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, u model.URL) (*model.URL, error) {
			return &u, nil
		})

	result, err := svc.ShortenURL(context.Background(), model.ShortenURLParams{
		LongURL: "https://example.com/very-long-path",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.LongURL != "https://example.com/very-long-path" {
		t.Fatalf("got long URL %q, want %q", result.LongURL, "https://example.com/very-long-path")
	}
	if result.ShortCode == "" {
		t.Fatal("expected non-empty short code")
	}
	if result.ID == "" {
		t.Fatal("expected non-empty ID")
	}
}

func TestShortenURL_CustomAlias(t *testing.T) {
	ctrl := gomock.NewController(t)
	urlRepo := mock.NewMockURLRepository(ctrl)
	svc := New(urlRepo, nil, nil, nil, nil, nil, nil)

	urlRepo.EXPECT().
		ShortCodeExists(gomock.Any(), "myalias").
		Return(false, nil)

	urlRepo.EXPECT().
		Create(gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, u model.URL) (*model.URL, error) {
			return &u, nil
		})

	result, err := svc.ShortenURL(context.Background(), model.ShortenURLParams{
		LongURL: "https://example.com",
		CustomAlias: "myalias",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ShortCode != "myalias" {
		t.Fatalf("got short code %q, want %q", result.ShortCode, "myalias")
	}
}

func TestShortenURL_CustomAliasTaken(t *testing.T) {
	ctrl := gomock.NewController(t)
	urlRepo := mock.NewMockURLRepository(ctrl)
	svc := New(urlRepo, nil, nil, nil, nil, nil, nil)

	urlRepo.EXPECT().
		ShortCodeExists(gomock.Any(), "taken").
		Return(true, nil)

	_, err := svc.ShortenURL(context.Background(), model.ShortenURLParams{
		LongURL: "https://example.com",
		CustomAlias: "taken",
	})
	if !errors.Is(err, model.ErrShortCodeTaken) {
		t.Fatalf("got %v, want ErrShortCodeTaken", err)
	}
}

func TestShortenURL_WithUserIDAndExpiry(t *testing.T) {
	ctrl := gomock.NewController(t)
	urlRepo := mock.NewMockURLRepository(ctrl)
	svc := New(urlRepo, nil, nil, nil, nil, nil, nil)

	urlRepo.EXPECT().
		ShortCodeExists(gomock.Any(), gomock.Any()).
		Return(false, nil)

	urlRepo.EXPECT().
		Create(gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, u model.URL) (*model.URL, error) {
			return &u, nil
		})

	userID := "user-123"
	expiry := time.Now().Add(24 * time.Hour)

	result, err := svc.ShortenURL(context.Background(), model.ShortenURLParams{
		LongURL: "https://example.com",
		UserID:      &userID,
		ExpiresAt:   &expiry,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.UserID == nil || *result.UserID != userID {
		t.Fatalf("got user ID %v, want %q", result.UserID, userID)
	}
	if result.ExpiresAt == nil {
		t.Fatal("expected non-nil ExpiresAt")
	}
}

func TestShortenURL_CollisionRetry(t *testing.T) {
	ctrl := gomock.NewController(t)
	urlRepo := mock.NewMockURLRepository(ctrl)
	svc := New(urlRepo, nil, nil, nil, nil, nil, nil)

	// First 2 generated codes collide, third succeeds.
	gomock.InOrder(
		urlRepo.EXPECT().ShortCodeExists(gomock.Any(), gomock.Any()).Return(true, nil),
		urlRepo.EXPECT().ShortCodeExists(gomock.Any(), gomock.Any()).Return(true, nil),
		urlRepo.EXPECT().ShortCodeExists(gomock.Any(), gomock.Any()).Return(false, nil),
	)

	urlRepo.EXPECT().
		Create(gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, u model.URL) (*model.URL, error) {
			return &u, nil
		})

	result, err := svc.ShortenURL(context.Background(), model.ShortenURLParams{
		LongURL: "https://example.com",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ShortCode == "" {
		t.Fatal("expected non-empty short code")
	}
}

func TestShortenURL_AllCollisionsExhausted(t *testing.T) {
	ctrl := gomock.NewController(t)
	urlRepo := mock.NewMockURLRepository(ctrl)
	svc := New(urlRepo, nil, nil, nil, nil, nil, nil)

	// All 5 attempts collide.
	urlRepo.EXPECT().
		ShortCodeExists(gomock.Any(), gomock.Any()).
		Return(true, nil).
		Times(maxGenerateAttempts)

	_, err := svc.ShortenURL(context.Background(), model.ShortenURLParams{
		LongURL: "https://example.com",
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestShortenURL_CreateError(t *testing.T) {
	ctrl := gomock.NewController(t)
	urlRepo := mock.NewMockURLRepository(ctrl)
	svc := New(urlRepo, nil, nil, nil, nil, nil, nil)

	urlRepo.EXPECT().
		ShortCodeExists(gomock.Any(), gomock.Any()).
		Return(false, nil)

	urlRepo.EXPECT().
		Create(gomock.Any(), gomock.Any()).
		Return(nil, errors.New("db connection lost"))

	_, err := svc.ShortenURL(context.Background(), model.ShortenURLParams{
		LongURL: "https://example.com",
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
