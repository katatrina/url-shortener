package service

import (
	"context"
	"log/slog"
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
			slog.Warn("cache get failed", "short_code", shortCode, "error", err)
		} else {
			if cachedURL != nil {
				// Cache HIT — URL found in Redis.
				metrics.CacheRequests.WithLabelValues("hit").Inc()

				if cachedURL.ExpiresAt != nil && time.Now().Unix() > *cachedURL.ExpiresAt {
					return "", model.ErrURLExpired
				}

				s.trackClick(cachedURL.ID, meta)
				return cachedURL.LongURL, nil
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
			LongURL: u.LongURL,
		}
		if u.ExpiresAt != nil {
			ts := u.ExpiresAt.Unix()
			cachedURL.ExpiresAt = &ts
		}

		if err := s.urlCache.Set(ctx, shortCode, cachedURL); err != nil {
			slog.Warn("cache set failed", "short_code", shortCode, "error", err)
		}
	}

	s.trackClick(u.ID, meta)

	return u.LongURL, nil
}

func (s *Service) trackClick(urlID string, meta model.ClickMeta) {
	if s.clickCollector == nil {
		return
	}

	s.clickCollector.Track(urlID, meta)
}
