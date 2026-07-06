package link

import "strings"

type CreateLinkRequest struct {
	DestinationURL string  `json:"destinationUrl" validate:"required,http_url,max=2048"`
	Title          *string `json:"title" validate:"omitempty,max=255"`
}

func (r *CreateLinkRequest) Normalize() {
	r.DestinationURL = strings.TrimSpace(r.DestinationURL)
}

type LinkResponse struct {
	ID             string  `json:"id"`
	UserID         string  `json:"userId"`
	Slug           string  `json:"slug"`
	DestinationURL string  `json:"destinationUrl"`
	Title          *string `json:"title"`
	IsCustomSlug   bool    `json:"isCustomSlug"`
	CreatedAt      int64   `json:"createdAt"`
	UpdatedAt      int64   `json:"updatedAt"`
}

func newLinkResponse(l *Link) LinkResponse {
	return LinkResponse{
		ID:             l.ID,
		UserID:         l.UserID,
		Slug:           l.Slug,
		DestinationURL: l.DestinationURL,
		Title:          l.Title,
		IsCustomSlug:   l.IsCustomSlug,
		CreatedAt:      l.CreatedAt.Unix(),
		UpdatedAt:      l.UpdatedAt.Unix(),
	}
}
