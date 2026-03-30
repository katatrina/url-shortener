package service

import (
	"context"
	"time"

	"github.com/katatrina/url-shortener/internal/model"
)

const (
	defaultTopLimit = 10
)

func (s *Service) GetURLStats(ctx context.Context, params model.GetURLStatsParams) (*model.URLStatsResponse, error) {
	u, err := s.urlRepo.FindByShortCode(ctx, params.ShortCode)
	if err != nil {
		return nil, err
	}

	if u.UserID == nil || *u.UserID != params.UserID {
		return nil, model.ErrURLOwnerMismatch
	}

	to := time.Now().UTC().Truncate(24 * time.Hour)
	from := to.AddDate(0, 0, -params.Days)

	dailyStats, err := s.statsRepo.GetDailyStats(ctx, u.ID, from, to)
	if err != nil {
		return nil, err
	}

	topReferrers, err := s.clickEventRepo.GetTopReferrers(ctx, u.ID, from, to, defaultTopLimit)
	if err != nil {
		return nil, err
	}

	topCountries, err := s.clickEventRepo.GetTopCountries(ctx, u.ID, from, to, defaultTopLimit)
	if err != nil {
		return nil, err
	}

	dailyResponses := make([]model.DailyStatResponse, len(dailyStats))
	for i, ds := range dailyStats {
		dailyResponses[i] = model.DailyStatResponse{
			Date:       ds.Date.Unix(),
			ClickCount: ds.ClickCount,
		}
	}

	return &model.URLStatsResponse{
		TotalClicks:  u.ClickCount,
		DailyClicks:  dailyResponses,
		TopReferrers: topReferrers,
		TopCountries: topCountries,
	}, nil
}
