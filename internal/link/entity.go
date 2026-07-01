package link

import "time"

type Link struct {
	ID             string    `db:"id"`
	OwnerID        string    `db:"owner_id"`
	Slug           string    `db:"slug"`
	DestinationURL string    `db:"destination_url"`
	Title          *string   `db:"title"`
	IsCustomSlug   bool      `db:"is_custom_slug"`
	CreatedAt      time.Time `db:"created_at"`
	UpdatedAt      time.Time `db:"updated_at"`
}
