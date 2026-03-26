package service

import (
	"context"
	"log"
	"time"

	"github.com/katatrina/url-shortener/internal/cache"
	"github.com/katatrina/url-shortener/internal/model"
)

// Resolve looks up a short code and returns the original URL and its internal ID.
//
// Phase 3 change: this method no longer tracks clicks.
// Click tracking is now handled by the analytics collector in the handler layer,
// because tracking requires HTTP metadata (IP, User-Agent, Referer) that the
// service layer shouldn't know about.
//
// Returns (originalURL, urlID, error).
// The caller needs urlID to create a ClickEvent for the analytics pipeline.
func (s *Service) Resolve(ctx context.Context, shortCode string) (string, string, error) {
	// Step 1: Try cache first
	if s.urlCache != nil {
		cachedURL, err := s.urlCache.Get(ctx, shortCode)
		if err != nil {
			log.Printf("[WARN] cache get failed for %s: %v", shortCode, err)
		} else {
			if cachedURL != nil {
				if cachedURL.ExpiresAt != nil && time.Now().Unix() > *cachedURL.ExpiresAt {
					return "", "", model.ErrURLExpired
				}

				return cachedURL.OriginalURL, cachedURL.ID, nil
			}
		}
	}

	// Step 2: Cache MISS (or cache disabled) - query DB.
	u, err := s.urlRepo.FindByShortCode(ctx, shortCode)
	if err != nil {
		return "", "", err
	}

	if u.IsExpired() {
		return "", "", model.ErrURLExpired
	}

	// Step 3: Populate cache for next time.
	if s.urlCache != nil {
		cachedURL := &cache.CachedURL{
			ID:          u.ID,
			OriginalURL: u.OriginalURL,
		}
		if u.ExpiresAt != nil {
			ts := u.ExpiresAt.Unix()
			cachedURL.ExpiresAt = &ts
		}

		if err := s.urlCache.Set(ctx, shortCode, cachedURL); err != nil {
			log.Printf("[WARN] cache set failed for %s: %v", shortCode, err)
		}
	}

	return u.OriginalURL, u.ID, nil
}
