package link

import (
	"strings"
	"time"
)

type CreateLinkRequest struct {
	DestinationURL string  `json:"destinationUrl" validate:"required,http_url,max=2048"`
	Title          string  `json:"title" validate:"omitempty,max=255"`
	Slug           *string `json:"slug" validate:"omitnil,min=3,max=30,slug"`
}

func (r *CreateLinkRequest) Normalize() {
	r.DestinationURL = strings.TrimSpace(r.DestinationURL)
	r.Title = strings.TrimSpace(r.Title)
	if r.Slug != nil {
		r.Slug = new(strings.TrimSpace(*r.Slug))
	}
}

type LinkResponse struct {
	ID             string    `json:"id"`
	UserID         string    `json:"userId"`
	Slug           string    `json:"slug"`
	ShortURL       string    `json:"shortUrl"`
	DestinationURL string    `json:"destinationUrl"`
	Title          string    `json:"title"`
	IsCustomSlug   bool      `json:"isCustomSlug"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
}

type LinkListItemResponse struct {
	LinkResponse
	Clicks int64 `json:"clicks"`
}

type ListLinksResponse struct {
	Items []LinkListItemResponse `json:"items"`
}

const (
	RangeLast24Hours = "24h"
	RangeLast7Days   = "7d"
	RangeLast30Days  = "30d"
	RangeLast90Days  = "90d"
)

const defaultStatsRange = RangeLast7Days

const maxTimezoneLen = 64

type StatsSummaryResponse struct {
	// TotalClicks is all-time and ignores the selected range.
	TotalClicks   int64 `json:"totalClicks"`
	ClicksInRange int64 `json:"clicksInRange"`
}

type TimeseriesPointResponse struct {
	// Timestamp is the start of the bucket, in UTC. Buckets with no clicks are
	// present with clicks: 0 — the series is dense.
	Timestamp time.Time `json:"timestamp"`
	Clicks    int64     `json:"clicks"`
}

// DimensionCountResponse is one row of a breakdown. An empty value means the
// dimension is unknown: no GeoIP match for a country, no referrer for direct
// traffic. Labelling that ("Unknown", "Direct") is a presentation decision.
type DimensionCountResponse struct {
	Value  string `json:"value"`
	Clicks int64  `json:"clicks"`
}

type LinkStatsResponse struct {
	LinkID      string    `json:"linkId"`
	Range       string    `json:"range"`
	Timezone    string    `json:"timezone"`
	Granularity string    `json:"granularity"`
	From        time.Time `json:"from"`
	To          time.Time `json:"to"`

	Summary      StatsSummaryResponse      `json:"summary"`
	Timeseries   []TimeseriesPointResponse `json:"timeseries"`
	TopCountries []DimensionCountResponse  `json:"topCountries"`
	TopReferrers []DimensionCountResponse  `json:"topReferrers"`
}

type UpdateLinkRequest struct {
	DestinationURL *string `json:"destinationUrl" validate:"omitnil,http_url,max=2048"`
	Title          *string `json:"title" validate:"omitnil,max=255"`
}

func (r *UpdateLinkRequest) Normalize() {
	if r.DestinationURL != nil {
		r.DestinationURL = new(strings.TrimSpace(*r.DestinationURL))
	}
	if r.Title != nil {
		r.Title = new(strings.TrimSpace(*r.Title))
	}
}

func (r *UpdateLinkRequest) IsEmpty() bool {
	return r.DestinationURL == nil && r.Title == nil
}

func newLinkResponse(l *Link, shortURLBase string) LinkResponse {
	return LinkResponse{
		ID:             l.ID,
		UserID:         l.UserID,
		Slug:           l.Slug,
		ShortURL:       shortURLBase + "/" + l.Slug,
		DestinationURL: l.DestinationURL,
		Title:          l.Title,
		IsCustomSlug:   l.IsCustomSlug,
		CreatedAt:      l.CreatedAt.UTC(),
		UpdatedAt:      l.UpdatedAt.UTC(),
	}
}

func newListLinksResponse(links []LinkListItem, shortURLBase string) ListLinksResponse {
	items := make([]LinkListItemResponse, 0, len(links))
	for i := range links {
		l := &links[i]

		items = append(items, LinkListItemResponse{
			LinkResponse: newLinkResponse(&l.Link, shortURLBase),
			Clicks:       l.ClickCount,
		})
	}
	return ListLinksResponse{Items: items}
}

func newLinkStatsResponse(s *LinkStats, rng string, loc *time.Location) LinkStatsResponse {
	timeseries := make([]TimeseriesPointResponse, 0, len(s.Stats.Timeseries))
	for _, p := range s.Stats.Timeseries {
		timeseries = append(timeseries, TimeseriesPointResponse{
			Timestamp: p.Bucket.UTC(),
			Clicks:    p.Clicks,
		})
	}

	granularity := BucketDay
	if rng == RangeLast24Hours {
		granularity = BucketHour
	}

	return LinkStatsResponse{
		LinkID:      s.Link.ID,
		Range:       rng,
		Timezone:    loc.String(),
		Granularity: granularity,
		From:        s.From.UTC(),
		To:          s.To.UTC(),
		Summary: StatsSummaryResponse{
			TotalClicks:   s.Stats.Summary.TotalClicks,
			ClicksInRange: s.Stats.Summary.ClicksInRange,
		},
		Timeseries:   timeseries,
		TopCountries: newDimensionCounts(s.Stats.TopCountries),
		TopReferrers: newDimensionCounts(s.Stats.TopReferrers),
	}
}

func newDimensionCounts(in []DimensionCount) []DimensionCountResponse {
	out := make([]DimensionCountResponse, 0, len(in))
	for _, d := range in {
		out = append(out, DimensionCountResponse{Value: d.Value, Clicks: d.Clicks})
	}
	return out
}
