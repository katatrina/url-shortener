package service

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/katatrina/url-shortener/internal/model"
	"github.com/katatrina/url-shortener/internal/shortcode"
)

const maxGenerateAttempts = 5

func (s *Service) ShortenURL(ctx context.Context, params model.ShortenURLParams) (*model.URL, error) {
	var code string

	if params.CustomAlias != "" {
		exists, err := s.urlRepo.ShortCodeExists(ctx, params.CustomAlias)
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, model.ErrShortCodeTaken
		}

		code = params.CustomAlias
	} else {
		var err error
		code, err = s.generateUniqueCode(ctx)
		if err != nil {
			return nil, err
		}
	}

	id, err := uuid.NewV7()
	if err != nil {
		return nil, fmt.Errorf("failed to generate URL ID: %w", err)
	}

	now := time.Now()
	newURL := model.URL{
		ID:          id.String(),
		ShortCode:   code,
		OriginalURL: params.OriginalURL,
		UserID:      params.UserID,
		ClickCount:  0,
		ExpiresAt:   params.ExpiresAt,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	created, err := s.urlRepo.Create(ctx, newURL)
	if err != nil {
		return nil, err
	}

	return created, nil
}

// generateUniqueCode generates a random short code and checks for collisions.
// Retries up to maxGenerateAttempts times.
func (s *Service) generateUniqueCode(ctx context.Context) (string, error) {
	for range maxGenerateAttempts {
		code := shortcode.Generate()

		exists, err := s.urlRepo.ShortCodeExists(ctx, code)
		if err != nil {
			return "", err
		}

		if !exists {
			return code, nil
		}

		// Collision - extremely rare, but retry.
	}

	return "", fmt.Errorf("failed to generate unique short code after %d attempts", maxGenerateAttempts)
}

func (s *Service) GetUserURL(ctx context.Context, shortCode, userID string) (*model.URL, error) {
	u, err := s.urlRepo.FindByShortCode(ctx, shortCode)
	if err != nil {
		return nil, err
	}

	if u.UserID == nil || *u.UserID != userID {
		return nil, model.ErrURLOwnerMismatch
	}

	return u, nil
}

func (s *Service) ListUserURLs(ctx context.Context, userID string, limit, offset int) ([]model.URL, int64, error) {
	urls, err := s.urlRepo.ListByUserID(ctx, userID, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	total, err := s.urlRepo.CountByUserID(ctx, userID)
	if err != nil {
		return nil, 0, err
	}

	return urls, total, nil
}

func (s *Service) DeleteUserURL(ctx context.Context, shortCode, userID string) error {
	u, err := s.urlRepo.FindByShortCode(ctx, shortCode)
	if err != nil {
		return err
	}

	if u.UserID == nil || *u.UserID != userID {
		return model.ErrURLOwnerMismatch
	}

	if err = s.urlRepo.Delete(ctx, u.ID); err != nil {
		return err
	}

	// Invalidate cache AFTER successful DB delete.
	// If DB delete fails, we don't want to remove a valid cache.
	// Worst case: cache delete fails, stale cache remains but will be auto-evicted by TTL.
	if s.urlCache != nil {
		if err := s.urlCache.Delete(ctx, shortCode); err != nil {
			log.Printf("[WARN] cache delete failed for %s: %v", shortCode, err)
		}
	}

	return nil
}
