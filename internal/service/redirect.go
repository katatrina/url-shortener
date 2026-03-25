package service

import (
	"context"
	"log"
	"time"

	"github.com/katatrina/url-shortener/internal/cache"
	"github.com/katatrina/url-shortener/internal/model"
)

// Resolve looks up a short code and returns the original URL for redirect.
func (s *Service) Resolve(ctx context.Context, shortCode string) (string, error) {
	// Step 1: Try cache first
	if s.urlCache != nil {
		cachedURL, err := s.urlCache.Get(ctx, shortCode)
		if err != nil {
			// Cache error (not cache MISS) - log and fall through DB.
			// Never let cache failure break the redirect flow.
			log.Printf("[WARN] cache get failed for %s: %v", shortCode, err)
		} else {
			if cachedURL != nil {
				// Cache HIT - check expiry and return.
				if cachedURL.ExpiresAt != nil && time.Now().Unix() > *cachedURL.ExpiresAt {
					return "", model.ErrURLExpired
				}

				go s.trackClick(context.WithoutCancel(ctx), cachedURL.ID, shortCode)

				return cachedURL.OriginalURL, nil
			}
		}
	}

	// Step 2: Cache MISS (or cache disabled) - query DB.
	u, err := s.urlRepo.FindByShortCode(ctx, shortCode)
	if err != nil {
		return "", err
	}

	if u.IsExpired() {
		return "", model.ErrURLExpired
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

		// Sync is fine here: cache write is sub-ms, not worth a goroutine.
		if err := s.urlCache.Set(ctx, shortCode, cachedURL); err != nil {
			// Cache write failure - log but don't fail the request.
			log.Printf("[WARN] cache set failed for %s: %v", shortCode, err)
		}
	}

	// Step 4: Try to increment click counter asynchronously (best effort).
	go s.trackClick(context.WithoutCancel(ctx), u.ID, shortCode)

	return u.OriginalURL, nil
}

// trackClick increments click count for a url.
// If error, it just logs it.
func (s *Service) trackClick(ctx context.Context, urlID, shortCode string) {
	if err := s.urlRepo.IncrementClickCount(ctx, urlID); err != nil {
		log.Printf("[WARN] failed to increment click count for %s: %v", shortCode, err)
	}
}
