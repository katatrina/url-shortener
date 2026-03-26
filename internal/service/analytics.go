package service

import (
	"context"
	"time"

	"github.com/katatrina/url-shortener/internal/model"
)

const (
	defaultStatsRangeDays = 30
	defaultTopLimit       = 10
)

// GetURLStats returns analytics data for a URL owned by the given user.
// Combines daily click stats, top referrers, and top countries.
func (s *Service) GetURLStats(ctx context.Context, shortCode, userID string) (*model.URLStats, error) {
	// Step 1: Verify ownership — same pattern as GetUserURL.
	u, err := s.urlRepo.FindByShortCode(ctx, shortCode)
	if err != nil {
		return nil, err
	}

	if u.UserID == nil || *u.UserID != userID {
		return nil, model.ErrURLOwnerMismatch
	}

	// Step 2: Fetch analytics data.
	to := time.Now().UTC().Truncate(24 * time.Hour)
	from := to.AddDate(0, 0, -defaultStatsRangeDays)

	dailyStats, err := s.statsRepo.GetDailyStats(ctx, u.ID, from, to)
	if err != nil {
		return nil, err
	}

	topReferrers, err := s.clickEventRepo.GetTopReferrers(ctx, u.ID, defaultTopLimit)
	if err != nil {
		return nil, err
	}

	topCountries, err := s.clickEventRepo.GetTopCountries(ctx, u.ID, defaultTopLimit)
	if err != nil {
		return nil, err
	}

	return &model.URLStats{
		DailyClicks:  dailyStats,
		TopReferrers: topReferrers,
		TopCountries: topCountries,
	}, nil
}
