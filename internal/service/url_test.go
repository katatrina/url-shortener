package service

import (
	"context"
	"errors"
	"testing"

	"github.com/katatrina/url-shortener/internal/mock"
	"github.com/katatrina/url-shortener/internal/model"
	"go.uber.org/mock/gomock"
)

// --- ShortenURL ---

func TestShortenURL_RandomCode_Success(t *testing.T) {
	ctrl := gomock.NewController(t)

	mockURLRepo := mock.NewMockURLRepository(ctrl)
	svc := New(mockURLRepo, nil, nil, nil)

	// The generated code is random, so we accept any string.
	mockURLRepo.EXPECT().
		ShortCodeExists(gomock.Any(), gomock.Any()).
		Return(false, nil)

	mockURLRepo.EXPECT().
		Create(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, url model.URL) (*model.URL, error) {
			if url.OriginalURL != "https://example.com/long" {
				t.Errorf("expected original URL https://example.com/long, got %s", url.OriginalURL)
			}
			if len(url.ShortCode) == 0 {
				t.Error("expected short code to be generated")
			}
			return &url, nil
		})

	result, err := svc.ShortenURL(context.Background(), model.ShortenURLParams{
		OriginalURL: "https://example.com/long",
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result.OriginalURL != "https://example.com/long" {
		t.Errorf("expected original URL in result, got %s", result.OriginalURL)
	}
}

func TestShortenURL_CustomAlias_Success(t *testing.T) {
	ctrl := gomock.NewController(t)

	mockURLRepo := mock.NewMockURLRepository(ctrl)
	svc := New(mockURLRepo, nil, nil, nil)

	mockURLRepo.EXPECT().
		ShortCodeExists(gomock.Any(), "myalias").
		Return(false, nil) // Alias is available.

	mockURLRepo.EXPECT().
		Create(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, url model.URL) (*model.URL, error) {
			if url.ShortCode != "myalias" {
				t.Errorf("expected short code myalias, got %s", url.ShortCode)
			}
			return &url, nil
		})

	result, err := svc.ShortenURL(context.Background(), model.ShortenURLParams{
		OriginalURL: "https://example.com",
		CustomAlias: "myalias",
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result.ShortCode != "myalias" {
		t.Errorf("expected short code myalias, got %s", result.ShortCode)
	}
}

func TestShortenURL_CustomAliasTaken(t *testing.T) {
	ctrl := gomock.NewController(t)

	mockURLRepo := mock.NewMockURLRepository(ctrl)
	svc := New(mockURLRepo, nil, nil, nil)

	mockURLRepo.EXPECT().
		ShortCodeExists(gomock.Any(), "taken").
		Return(true, nil) // Already taken.

	_, err := svc.ShortenURL(context.Background(), model.ShortenURLParams{
		OriginalURL: "https://example.com",
		CustomAlias: "taken",
	})

	if !errors.Is(err, model.ErrShortCodeTaken) {
		t.Errorf("expected ErrShortCodeTaken, got %v", err)
	}
}

func TestShortenURL_CollisionThenSuccess(t *testing.T) {
	// First generated code collides, second attempt succeeds.
	// This tests the retry loop in generateUniqueCode.
	ctrl := gomock.NewController(t)

	mockURLRepo := mock.NewMockURLRepository(ctrl)
	svc := New(mockURLRepo, nil, nil, nil)

	callCount := 0
	mockURLRepo.EXPECT().
		ShortCodeExists(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, code string) (bool, error) {
			callCount++
			if callCount == 1 {
				return true, nil // First code: collision!
			}
			return false, nil // Second code: available.
		}).
		Times(2)

	mockURLRepo.EXPECT().
		Create(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, url model.URL) (*model.URL, error) {
			return &url, nil
		})

	_, err := svc.ShortenURL(context.Background(), model.ShortenURLParams{
		OriginalURL: "https://example.com",
	})

	if err != nil {
		t.Fatalf("expected no error after retry, got %v", err)
	}
}

func TestShortenURL_WithAuthenticatedUser(t *testing.T) {
	ctrl := gomock.NewController(t)

	mockURLRepo := mock.NewMockURLRepository(ctrl)
	svc := New(mockURLRepo, nil, nil, nil)

	userID := "user-123"

	mockURLRepo.EXPECT().
		ShortCodeExists(gomock.Any(), gomock.Any()).
		Return(false, nil)

	mockURLRepo.EXPECT().
		Create(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, url model.URL) (*model.URL, error) {
			if url.UserID == nil {
				t.Error("expected user ID to be set")
			} else if *url.UserID != "user-123" {
				t.Errorf("expected user ID user-123, got %s", *url.UserID)
			}
			return &url, nil
		})

	_, err := svc.ShortenURL(context.Background(), model.ShortenURLParams{
		OriginalURL: "https://example.com",
		UserID:      &userID,
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

// --- GetUserURL ---

func TestGetUserURL_Success(t *testing.T) {
	ctrl := gomock.NewController(t)

	mockURLRepo := mock.NewMockURLRepository(ctrl)
	svc := New(mockURLRepo, nil, nil, nil)

	userID := "user-123"
	storedURL := &model.URL{
		ID:          "url-456",
		ShortCode:   "aB3kX9m",
		OriginalURL: "https://example.com",
		UserID:      &userID,
	}

	mockURLRepo.EXPECT().
		FindByShortCode(gomock.Any(), "aB3kX9m").
		Return(storedURL, nil)

	url, err := svc.GetUserURL(context.Background(), "aB3kX9m", "user-123")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if url.ID != "url-456" {
		t.Errorf("expected url ID url-456, got %s", url.ID)
	}
}

func TestGetUserURL_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)

	mockURLRepo := mock.NewMockURLRepository(ctrl)
	svc := New(mockURLRepo, nil, nil, nil)

	mockURLRepo.EXPECT().
		FindByShortCode(gomock.Any(), "nonexist").
		Return(nil, model.ErrURLNotFound)

	_, err := svc.GetUserURL(context.Background(), "nonexist", "user-123")

	if !errors.Is(err, model.ErrURLNotFound) {
		t.Errorf("expected ErrURLNotFound, got %v", err)
	}
}

func TestGetUserURL_OwnerMismatch(t *testing.T) {
	// User A tries to access User B's URL → should be rejected.
	ctrl := gomock.NewController(t)

	mockURLRepo := mock.NewMockURLRepository(ctrl)
	svc := New(mockURLRepo, nil, nil, nil)

	ownerID := "user-999" // Owner is user-999.
	storedURL := &model.URL{
		ID:     "url-456",
		UserID: &ownerID,
	}

	mockURLRepo.EXPECT().
		FindByShortCode(gomock.Any(), "aB3kX9m").
		Return(storedURL, nil)

	// Requesting user is user-123, not user-999.
	_, err := svc.GetUserURL(context.Background(), "aB3kX9m", "user-123")

	if !errors.Is(err, model.ErrURLOwnerMismatch) {
		t.Errorf("expected ErrURLOwnerMismatch, got %v", err)
	}
}

// --- ListUserURLs ---

func TestListUserURLs_Success(t *testing.T) {
	ctrl := gomock.NewController(t)

	mockURLRepo := mock.NewMockURLRepository(ctrl)
	svc := New(mockURLRepo, nil, nil, nil)

	storedURLs := []model.URL{
		{ID: "url-1", ShortCode: "code1", OriginalURL: "https://example.com/1"},
		{ID: "url-2", ShortCode: "code2", OriginalURL: "https://example.com/2"},
	}

	mockURLRepo.EXPECT().
		ListByUserID(gomock.Any(), "user-123", 10, 0).
		Return(storedURLs, nil)

	mockURLRepo.EXPECT().
		CountByUserID(gomock.Any(), "user-123").
		Return(int64(2), nil)

	urls, total, err := svc.ListUserURLs(context.Background(), "user-123", 10, 0)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(urls) != 2 {
		t.Errorf("expected 2 URLs, got %d", len(urls))
	}
	if total != 2 {
		t.Errorf("expected total 2, got %d", total)
	}
}

// --- DeleteUserURL ---

func TestDeleteUserURL_Success(t *testing.T) {
	ctrl := gomock.NewController(t)

	mockURLRepo := mock.NewMockURLRepository(ctrl)
	svc := New(mockURLRepo, nil, nil, nil)

	userID := "user-123"
	storedURL := &model.URL{
		ID:     "url-456",
		UserID: &userID,
	}

	mockURLRepo.EXPECT().
		FindByShortCode(gomock.Any(), "aB3kX9m").
		Return(storedURL, nil)

	mockURLRepo.EXPECT().
		Delete(gomock.Any(), "url-456").
		Return(nil)

	err := svc.DeleteUserURL(context.Background(), "aB3kX9m", "user-123")

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestDeleteUserURL_OwnerMismatch(t *testing.T) {
	ctrl := gomock.NewController(t)

	mockURLRepo := mock.NewMockURLRepository(ctrl)
	svc := New(mockURLRepo, nil, nil, nil)

	ownerID := "user-999"
	storedURL := &model.URL{
		ID:     "url-456",
		UserID: &ownerID,
	}

	mockURLRepo.EXPECT().
		FindByShortCode(gomock.Any(), "aB3kX9m").
		Return(storedURL, nil)

	// Delete should NOT be called — ownership check fails first.

	err := svc.DeleteUserURL(context.Background(), "aB3kX9m", "user-123")

	if !errors.Is(err, model.ErrURLOwnerMismatch) {
		t.Errorf("expected ErrURLOwnerMismatch, got %v", err)
	}
}
