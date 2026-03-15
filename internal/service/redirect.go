package service

import (
	"context"
	"log"

	"github.com/katatrina/url-shortener/internal/model"
)

// Resolve looks up a short code and returns the original URL for redirect.
// It also increments the click counter synchronously.
func (s *Service) Resolve(ctx context.Context, shortCode string) (string, error) {
	u, err := s.urlRepo.FindByShortCode(ctx, shortCode)
	if err != nil {
		return "", err
	}

	if u.IsExpired() {
		return "", model.ErrURLExpired
	}

	// For now, we implement synchronous click tracking.
	// Later this will be replaced by pushing to an async queue.
	if err = s.urlRepo.IncrementClickCount(ctx, u.ID); err != nil {
		// Click tracking failure should NOT block the redirect.
		// User experience > analytics accuracy.
		log.Printf("[WARN] failed to increment click count for %s: %v", shortCode, err)
	}

	return u.OriginalURL, nil
}
