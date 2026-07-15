package link

import "strings"

type CreateLinkRequest struct {
	DestinationURL string  `json:"destinationUrl" validate:"required,http_url,max=2048"`
	Title          string  `json:"title" validate:"omitempty,max=255"`
	Slug           *string `json:"slug" validate:"omitempty,min=3,max=30,slug"`
}

func (r *CreateLinkRequest) Normalize() {
	r.DestinationURL = strings.TrimSpace(r.DestinationURL)
	if r.Slug != nil {
		r.Slug = new(strings.TrimSpace(*r.Slug))
	}
}

type LinkResponse struct {
	ID             string  `json:"id"`
	UserID         string  `json:"userId"`
	Slug           string  `json:"slug"`
	ShortURL       string  `json:"shortUrl"`
	DestinationURL string  `json:"destinationUrl"`
	Title          string  `json:"title"`
	IsCustomSlug   bool    `json:"isCustomSlug"`
	CreatedAt      int64   `json:"createdAt"`
	UpdatedAt      int64   `json:"updatedAt"`
}

type ListLinksResponse struct {
	Items []LinkResponse `json:"items"`
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
		CreatedAt:      l.CreatedAt.Unix(),
		UpdatedAt:      l.UpdatedAt.Unix(),
	}
}

func newListLinksResponse(links []Link, shortURLBase string) ListLinksResponse {
	items := make([]LinkResponse, 0, len(links))
	for i := range links {
		items = append(items, newLinkResponse(&links[i], shortURLBase))
	}
	return ListLinksResponse{Items: items}
}
