package link

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/katatrina/url-shortener/internal/slug"
)

const maxSlugRetries = 3

const topDimensionLimit = 5

type Service struct {
	linkRepo        *Repository
	maxLinksPerUser int
}

func NewService(linkRepo *Repository, maxLinksPerUser int) *Service {
	return &Service{
		linkRepo:        linkRepo,
		maxLinksPerUser: maxLinksPerUser,
	}
}

type CreateLinkParams struct {
	UserID         string
	DestinationURL string
	Title          string
	Slug           *string
}

func (s *Service) CreateLink(ctx context.Context, arg CreateLinkParams) (*Link, error) {
	// There is a very little chance of race condition here. But it's fine.
	count, err := s.linkRepo.Count(ctx, arg.UserID)
	if err != nil {
		return nil, fmt.Errorf("count user links: %w", err)
	}
	if count >= int64(s.maxLinksPerUser) {
		return nil, ErrLinkQuotaExceeded
	}

	if arg.Slug != nil {
		return s.createWithCustomSlug(ctx, arg)
	}
	return s.createWithGeneratedSlug(ctx, arg)
}

func (s *Service) createWithGeneratedSlug(ctx context.Context, arg CreateLinkParams) (*Link, error) {
	for attempt := range maxSlugRetries {
		generatedSlug := slug.Generate()
		id, _ := uuid.NewV7()

		link, err := s.linkRepo.Insert(ctx, InsertLinkCommand{
			ID:             id.String(),
			UserID:         arg.UserID,
			Slug:           generatedSlug,
			DestinationURL: arg.DestinationURL,
			Title:          arg.Title,
			IsCustomSlug:   false,
		})
		if err != nil {
			if errors.Is(err, ErrSlugExists) {
				slog.WarnContext(ctx, "slug collision, retrying",
					slog.String("slug", generatedSlug),
					slog.Int("attempt", attempt+1),
				)
				continue
			}
			return nil, fmt.Errorf("insert link: %w", err)
		}

		return link, nil
	}

	return nil, fmt.Errorf("generate unique slug: %d retries exhausted", maxSlugRetries)
}

func (s *Service) createWithCustomSlug(ctx context.Context, arg CreateLinkParams) (*Link, error) {
	id, _ := uuid.NewV7()
	link, err := s.linkRepo.Insert(ctx, InsertLinkCommand{
		ID:             id.String(),
		UserID:         arg.UserID,
		Slug:           *arg.Slug,
		DestinationURL: arg.DestinationURL,
		Title:          arg.Title,
		IsCustomSlug:   true,
	})
	if err != nil {
		return nil, fmt.Errorf("insert link: %w", err)
	}

	return link, nil
}

func (s *Service) ResolveSlug(ctx context.Context, slug string) (*LinkRow, error) {
	row, err := s.linkRepo.FindRowBySlug(ctx, slug)
	if err != nil {
		return nil, fmt.Errorf("find link row by slug: %w", err)
	}

	return row, nil
}

func (s *Service) ListLinks(ctx context.Context, userID string) ([]Link, error) {
	links, err := s.linkRepo.List(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("list links: %w", err)
	}

	return links, nil
}

func (s *Service) GetLink(ctx context.Context, id, userID string) (*Link, error) {
	link, err := s.linkRepo.FindByIDAndUserID(ctx, id, userID)
	if err != nil {
		return nil, fmt.Errorf("find link: %w", err)
	}

	return link, nil
}

type GetLinkStatsParams struct {
	LinkID   string
	UserID   string
	Range    string
	Location *time.Location
}

type LinkStats struct {
	LinkID       string
	From         time.Time
	To           time.Time
	PreviousFrom time.Time
	PreviousTo   time.Time
	Granularity  string

	Clicks         int64
	PreviousClicks *int64

	Timeseries   []TimePoint
	TopCountries []DimensionCount
	TopReferrers []DimensionCount
}

func (s *Service) GetLinkStats(ctx context.Context, arg GetLinkStatsParams) (*LinkStats, error) {
	row, err := s.linkRepo.FindRowByIDAndUserID(ctx, arg.LinkID, arg.UserID)
	if err != nil {
		return nil, fmt.Errorf("find link row: %w", err)
	}

	window, ok := newStatsWindow(arg.Range, arg.Location)
	if !ok {
		return nil, fmt.Errorf("unsupported stats range %q", arg.Range)
	}

	clickStats, err := s.linkRepo.AggregateClicks(ctx, ClickStatsQuery{
		LinkID:       arg.LinkID,
		From:         window.From,
		To:           window.To,
		PreviousFrom: window.PreviousFrom,
		PreviousTo:   window.PreviousTo,
		Bucket:       window.Bucket,
		Timezone:     arg.Location.String(),
		TopN:         topDimensionLimit,
	})
	if err != nil {
		return nil, fmt.Errorf("aggregate clicks: %w", err)
	}

	linkStats := &LinkStats{
		LinkID:       clickStats.LinkID,
		From:         window.From,
		To:           window.To,
		PreviousFrom: window.PreviousFrom,
		PreviousTo:   window.PreviousTo,
		Granularity:  window.Bucket,
		Clicks:       clickStats.Summary.Clicks,
		Timeseries:   clickStats.Timeseries,
		TopCountries: clickStats.TopCountries,
		TopReferrers: clickStats.TopReferrers,
	}
	if !row.CreatedAt.After(window.PreviousFrom) {
		linkStats.PreviousClicks = &clickStats.Summary.PreviousClicks
	}

	return linkStats, nil
}

type statsWindow struct {
	From, To                 time.Time
	PreviousFrom, PreviousTo time.Time
	Bucket                   string
}

func newStatsWindow(rng string, loc *time.Location) (statsWindow, bool) {
	now := time.Now().In(loc)

	switch rng {
	case RangeLast24Hours:
		startOfHour := time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), 0, 0, 0, loc)
		from := startOfHour.Add(-23 * time.Hour)
		return statsWindow{
			From:         from,
			To:           now,
			PreviousFrom: from.Add(-24 * time.Hour),
			PreviousTo:   now.Add(-24 * time.Hour),
			Bucket:       BucketHour,
		}, true
	case RangeLast7Days:
		return newDayStatsWindow(now, 7, loc), true
	case RangeLast30Days:
		return newDayStatsWindow(now, 30, loc), true
	case RangeLast90Days:
		return newDayStatsWindow(now, 90, loc), true
	}

	return statsWindow{}, false
}

func newDayStatsWindow(now time.Time, days int, loc *time.Location) statsWindow {
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)
	from := startOfDay.AddDate(0, 0, -(days - 1))

	return statsWindow{
		From:         from,
		To:           now,
		PreviousFrom: from.AddDate(0, 0, -days),
		PreviousTo:   now.AddDate(0, 0, -days),
		Bucket:       BucketDay,
	}
}

type UpdateLinkParams struct {
	ID             string
	UserID         string
	DestinationURL *string // nil = unchanged
	Title          *string // nil = unchanged, "" = clear
}

func (s *Service) UpdateLink(ctx context.Context, arg UpdateLinkParams) (*Link, error) {
	link, err := s.linkRepo.Update(ctx, UpdateLinkCommand(arg))
	if err != nil {
		return nil, fmt.Errorf("update link: %w", err)
	}

	return link, nil
}

func (s *Service) DeleteLink(ctx context.Context, id, userID string) error {
	if err := s.linkRepo.Delete(ctx, id, userID); err != nil {
		return fmt.Errorf("delete link: %w", err)
	}

	return nil
}
