package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/katatrina/url-shortener/internal/mock"
	"github.com/katatrina/url-shortener/internal/model"
	"go.uber.org/mock/gomock"
)

func TestResolve_Success(t *testing.T) {
	ctrl := gomock.NewController(t)

	mockURLRepo := mock.NewMockURLRepository(ctrl)
	svc := New(mockURLRepo, nil, nil, nil, nil, nil)

	storedURL := &model.URL{
		ID:          "url-123",
		ShortCode:   "aB3kX9m",
		OriginalURL: "https://example.com/very-long-url",
	}

	mockURLRepo.EXPECT().
		FindByShortCode(gomock.Any(), "aB3kX9m").
		Return(storedURL, nil)

	originalURL, urlID, err := svc.Resolve(context.Background(), "aB3kX9m")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if originalURL != "https://example.com/very-long-url" {
		t.Errorf("expected original URL, got %s", originalURL)
	}
	if urlID != "url-123" {
		t.Errorf("expected url ID url-123, got %s", urlID)
	}
}

func TestResolve_URLNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)

	mockURLRepo := mock.NewMockURLRepository(ctrl)
	svc := New(mockURLRepo, nil, nil, nil, nil, nil)

	mockURLRepo.EXPECT().
		FindByShortCode(gomock.Any(), "nonexist").
		Return(nil, model.ErrURLNotFound)

	_, _, err := svc.Resolve(context.Background(), "nonexist")

	if !errors.Is(err, model.ErrURLNotFound) {
		t.Errorf("expected ErrURLNotFound, got %v", err)
	}
}

func TestResolve_URLExpired(t *testing.T) {
	ctrl := gomock.NewController(t)

	mockURLRepo := mock.NewMockURLRepository(ctrl)
	svc := New(mockURLRepo, nil, nil, nil, nil, nil)

	expiredTime := time.Now().Add(-1 * time.Hour)
	storedURL := &model.URL{
		ID:          "url-123",
		ShortCode:   "expired1",
		OriginalURL: "https://example.com",
		ExpiresAt:   &expiredTime,
	}

	mockURLRepo.EXPECT().
		FindByShortCode(gomock.Any(), "expired1").
		Return(storedURL, nil)

	_, _, err := svc.Resolve(context.Background(), "expired1")

	if !errors.Is(err, model.ErrURLExpired) {
		t.Errorf("expected ErrURLExpired, got %v", err)
	}
}
