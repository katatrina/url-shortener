package link

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/katatrina/url-shortener/internal/click"
	"github.com/katatrina/url-shortener/internal/slug"
)

const maxSlugRetries = 3

// topDimensionLimit caps each top-N breakdown. Ten rows is what a dashboard
// panel can show without a scrollbar; a long tail of one-click referrers is
// noise, not insight.
const topDimensionLimit = 10

type ClickStatsReader interface {
	LinkStats(ctx context.Context, q click.StatsQuery) (*click.Stats, error)
}

type Service struct {
	linkRepo        *Repository
	clickStats      ClickStatsReader
	maxLinksPerUser int
}

func NewService(linkRepo *Repository, clickStats ClickStatsReader, maxLinksPerUser int) *Service {
	return &Service{
		linkRepo:        linkRepo,
		clickStats:      clickStats,
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
	count, err := s.linkRepo.CountByUserID(ctx, arg.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to count user links: %w", err)
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
	for n := range maxSlugRetries {
		generatedSlug := slug.Generate()
		id, _ := uuid.NewV7()

		link, err := s.linkRepo.Insert(ctx, InsertLinkParams{
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
					slog.Int("attempt", n+1),
				)
				continue
			}
			return nil, err
		}

		return link, nil
	}

	return nil, fmt.Errorf("failed to generate a unique slug after %d retries", maxSlugRetries)
}

func (s *Service) createWithCustomSlug(ctx context.Context, arg CreateLinkParams) (*Link, error) {
	id, _ := uuid.NewV7()
	return s.linkRepo.Insert(ctx, InsertLinkParams{
		ID:             id.String(),
		UserID:         arg.UserID,
		Slug:           *arg.Slug,
		DestinationURL: arg.DestinationURL,
		Title:          arg.Title,
		IsCustomSlug:   true,
	})
}

func (s *Service) ResolveSlug(ctx context.Context, rawSlug string) (*Link, error) {
	return s.linkRepo.FindBySlug(ctx, rawSlug)
}

func (s *Service) ListLinks(ctx context.Context, userID string) ([]LinkListItem, error) {
	return s.linkRepo.ListByUserID(ctx, userID)
}

type GetLinkStatsParams struct {
	LinkID   string
	UserID   string
	Range    string
	Location *time.Location
}

// LinkStats is everything the analytics screen renders in one shot.
type LinkStats struct {
	Link  *Link
	From  time.Time
	To    time.Time
	Stats *click.Stats
}

func (s *Service) GetLinkStats(ctx context.Context, arg GetLinkStatsParams) (*LinkStats, error) {
	link, err := s.linkRepo.FindByIDAndUserID(ctx, arg.LinkID, arg.UserID)
	if err != nil {
		return nil, err
	}

	from, bucket, ok := statsWindow(arg.Range, arg.Location)
	if !ok {
		return nil, fmt.Errorf("unsupported stats range %q", arg.Range)
	}

	stats, err := s.clickStats.LinkStats(ctx, click.StatsQuery{
		LinkID:   link.ID,
		From:     from,
		Bucket:   bucket,
		Timezone: arg.Location.String(),
		TopN:     topDimensionLimit,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to read link stats: %w", err)
	}

	return &LinkStats{
		Link:  link,
		From:  from,
		To:    time.Now(),
		Stats: stats,
	}, nil
}

// statsWindow turns a range keyword into a start instant aligned to a bucket
// boundary in loc, plus the bucket size.
//
// Alignment matters: "7d" means the last seven calendar days in the user's own
// timezone (today plus the six before it), not "168 hours ago", which would
// leave a half-empty bucket at each end of the chart. Ranges therefore always
// produce a fixed bucket count: 24, 7, 30, 90.
func statsWindow(rng string, loc *time.Location) (from time.Time, bucket string, ok bool) {
	now := time.Now().In(loc)

	startOfHour := func() time.Time {
		return time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), 0, 0, 0, loc)
	}
	startOfDay := func() time.Time {
		return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)
	}

	switch rng {
	case RangeLast24Hours:
		return startOfHour().Add(-23 * time.Hour), click.BucketHour, true
	case RangeLast7Days:
		return startOfDay().AddDate(0, 0, -6), click.BucketDay, true
	case RangeLast30Days:
		return startOfDay().AddDate(0, 0, -29), click.BucketDay, true
	case RangeLast90Days:
		return startOfDay().AddDate(0, 0, -89), click.BucketDay, true
	}

	return time.Time{}, "", false
}

type UpdateLinkParams struct {
	ID             string
	UserID         string
	DestinationURL *string // nil = unchanged
	Title          *string // nil = unchanged, "" = clear
}

func (s *Service) UpdateLink(ctx context.Context, arg UpdateLinkParams) (*Link, error) {
	return s.linkRepo.Update(ctx, arg)
}

func (s *Service) DeleteLink(ctx context.Context, id, userID string) error {
	return s.linkRepo.DeleteByIDAndUserID(ctx, id, userID)
}
