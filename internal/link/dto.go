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

type ListLinksResponse struct {
	Items []LinkResponse `json:"items"`
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

func newListLinksResponse(links []Link, shortURLBase string) ListLinksResponse {
	items := make([]LinkResponse, 0, len(links))
	for i := range links {
		items = append(items, newLinkResponse(&links[i], shortURLBase))
	}
	return ListLinksResponse{Items: items}
}
