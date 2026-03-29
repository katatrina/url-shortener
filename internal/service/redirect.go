package service

import (
	"context"
	"log"
	"time"

	"github.com/katatrina/url-shortener/internal/cache"
	"github.com/katatrina/url-shortener/internal/metrics"
	"github.com/katatrina/url-shortener/internal/model"
)

func (s *Service) Resolve(ctx context.Context, shortCode string, meta model.ClickMeta) (string, error) {
	if s.urlCache != nil {
		cachedURL, err := s.urlCache.Get(ctx, shortCode)
		if err != nil {
			// Redis error — not a miss, it's an infrastructure failure.
			metrics.CacheErrors.Inc()
			log.Printf("[WARN] cache get failed for %s: %v", shortCode, err)
		} else {
			if cachedURL != nil {
				// Cache HIT — URL found in Redis.
				metrics.CacheRequests.WithLabelValues("hit").Inc()

				if cachedURL.ExpiresAt != nil && time.Now().Unix() > *cachedURL.ExpiresAt {
					return "", model.ErrURLExpired
				}

				s.trackClick(cachedURL.ID, meta)
				return cachedURL.OriginalURL, nil
			}
			// Cache MISS — URL not in Redis, will query DB.
			metrics.CacheRequests.WithLabelValues("miss").Inc()
		}
	}

	u, err := s.urlRepo.FindByShortCode(ctx, shortCode)
	if err != nil {
		return "", err
	}

	if u.IsExpired() {
		return "", model.ErrURLExpired
	}

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

	s.trackClick(u.ID, meta)

	return u.OriginalURL, nil
}

func (s *Service) trackClick(urlID string, meta model.ClickMeta) {
	if s.clickCollector == nil {
		return
	}

	s.clickCollector.Track(urlID, meta)
}
