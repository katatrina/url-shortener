package service

import (
	"context"
	"time"

	"github.com/katatrina/url-shortener/internal/logger"

	"github.com/katatrina/url-shortener/internal/cache"
	"github.com/katatrina/url-shortener/internal/model"
)

func (s *Service) ResolveAndTrack(ctx context.Context, shortCode string, clickInfo model.ClickInfo) (string, error) {
	log := logger.FromRequestContext(ctx)
	if s.urlCache != nil {
		cachedURL, err := s.urlCache.Get(ctx, shortCode)
		if err != nil {
			// Redis error — not a miss, it's an infrastructure failure.
			log.Warn("cache get failed", "short_code", shortCode, "error", err)
		} else {
			if cachedURL != nil {
				if cachedURL.ExpiresAt != nil && time.Now().Unix() > *cachedURL.ExpiresAt {
					return "", model.ErrURLExpired
				}

				s.clickCollector.Track(cachedURL.ID, clickInfo)
				return cachedURL.LongURL, nil
			}
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
			ID:      u.ID,
			LongURL: u.LongURL,
		}
		if u.ExpiresAt != nil {
			ts := u.ExpiresAt.Unix()
			cachedURL.ExpiresAt = &ts
		}

		if err := s.urlCache.Set(ctx, shortCode, cachedURL); err != nil {
			log.Warn("cache set failed", "short_code", shortCode, "error", err)
		}
	}

	s.clickCollector.Track(u.ID, clickInfo)

	return u.LongURL, nil
}
