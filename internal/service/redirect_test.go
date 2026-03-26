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
	mockCollector := mock.NewMockClickCollector(ctrl)
	svc := New(mockURLRepo, nil, nil, nil, nil, nil, mockCollector)

	storedURL := &model.URL{
		ID:          "url-123",
		ShortCode:   "aB3kX9m",
		OriginalURL: "https://example.com/very-long-url",
	}

	mockURLRepo.EXPECT().
		FindByShortCode(gomock.Any(), "aB3kX9m").
		Return(storedURL, nil)

	mockCollector.EXPECT().
		Track(gomock.Any()) // Verify tracking was called

	originalURL, err := svc.Resolve(context.Background(), "aB3kX9m", model.ClickMeta{
		IP:        "192.168.1.1",
		UserAgent: "Mozilla/5.0",
		Referer:   "https://google.com",
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if originalURL != "https://example.com/very-long-url" {
		t.Errorf("expected original URL, got %s", originalURL)
	}
}

func TestResolve_URLNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)

	mockURLRepo := mock.NewMockURLRepository(ctrl)
	svc := New(mockURLRepo, nil, nil, nil, nil, nil, nil) // nil collector — won't be called

	mockURLRepo.EXPECT().
		FindByShortCode(gomock.Any(), "nonexist").
		Return(nil, model.ErrURLNotFound)

	_, err := svc.Resolve(context.Background(), "nonexist", model.ClickMeta{})

	if !errors.Is(err, model.ErrURLNotFound) {
		t.Errorf("expected ErrURLNotFound, got %v", err)
	}
}

func TestResolve_URLExpired(t *testing.T) {
	ctrl := gomock.NewController(t)

	mockURLRepo := mock.NewMockURLRepository(ctrl)
	svc := New(mockURLRepo, nil, nil, nil, nil, nil, nil)

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

	// Collector should NOT be called for expired URLs.

	_, err := svc.Resolve(context.Background(), "expired1", model.ClickMeta{})

	if !errors.Is(err, model.ErrURLExpired) {
		t.Errorf("expected ErrURLExpired, got %v", err)
	}
}
