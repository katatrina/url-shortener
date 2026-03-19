package service

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/katatrina/url-shortener/internal/mock"
	"github.com/katatrina/url-shortener/internal/model"
	"go.uber.org/mock/gomock"
)

func TestResolve_Success(t *testing.T) {
	ctrl := gomock.NewController(t)

	mockURLRepo := mock.NewMockURLRepository(ctrl)
	svc := New(mockURLRepo, nil, nil, nil) // added nil for urlCache

	storedURL := &model.URL{
		ID:          "url-123",
		ShortCode:   "aB3kX9m",
		OriginalURL: "https://example.com/very-long-url",
	}

	mockURLRepo.EXPECT().
		FindByShortCode(gomock.Any(), "aB3kX9m").
		Return(storedURL, nil)

	// IncrementClickCount now runs in a goroutine.
	// Use WaitGroup to ensure the goroutine completes before test cleanup.
	var wg sync.WaitGroup
	wg.Add(1)
	mockURLRepo.EXPECT().
		IncrementClickCount(gomock.Any(), "url-123").
		DoAndReturn(func(ctx context.Context, id string) error {
			defer wg.Done()
			return nil
		})

	result, err := svc.Resolve(context.Background(), "aB3kX9m")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result != "https://example.com/very-long-url" {
		t.Errorf("expected original URL, got %s", result)
	}

	wg.Wait() // Wait for async click tracking to complete.
}

func TestResolve_URLNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)

	mockURLRepo := mock.NewMockURLRepository(ctrl)
	svc := New(mockURLRepo, nil, nil, nil)

	mockURLRepo.EXPECT().
		FindByShortCode(gomock.Any(), "nonexist").
		Return(nil, model.ErrURLNotFound)

	// IncrementClickCount should NOT be called — URL doesn't exist.

	_, err := svc.Resolve(context.Background(), "nonexist")

	if !errors.Is(err, model.ErrURLNotFound) {
		t.Errorf("expected ErrURLNotFound, got %v", err)
	}
}

func TestResolve_URLExpired(t *testing.T) {
	ctrl := gomock.NewController(t)

	mockURLRepo := mock.NewMockURLRepository(ctrl)
	svc := New(mockURLRepo, nil, nil, nil)

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

	// IncrementClickCount should NOT be called for expired URLs.

	_, err := svc.Resolve(context.Background(), "expired1")

	if !errors.Is(err, model.ErrURLExpired) {
		t.Errorf("expected ErrURLExpired, got %v", err)
	}
}

func TestResolve_ClickCountFailure_StillRedirects(t *testing.T) {
	// This tests an important business decision:
	// user experience > analytics accuracy.
	ctrl := gomock.NewController(t)

	mockURLRepo := mock.NewMockURLRepository(ctrl)
	svc := New(mockURLRepo, nil, nil, nil)

	storedURL := &model.URL{
		ID:          "url-123",
		ShortCode:   "aB3kX9m",
		OriginalURL: "https://example.com",
	}

	mockURLRepo.EXPECT().
		FindByShortCode(gomock.Any(), "aB3kX9m").
		Return(storedURL, nil)

	var wg sync.WaitGroup
	wg.Add(1)
	mockURLRepo.EXPECT().
		IncrementClickCount(gomock.Any(), "url-123").
		DoAndReturn(func(ctx context.Context, id string) error {
			defer wg.Done()
			return context.DeadlineExceeded // DB timeout
		})

	// Should still succeed despite click tracking failure.
	result, err := svc.Resolve(context.Background(), "aB3kX9m")

	if err != nil {
		t.Fatalf("expected no error (redirect should work even if click tracking fails), got %v", err)
	}
	if result != "https://example.com" {
		t.Errorf("expected original URL, got %s", result)
	}

	wg.Wait()
}
