package link

import "time"

type Link struct {
	ID             string    `db:"id"`
	UserID         string    `db:"user_id"`
	Slug           string    `db:"slug"`
	DestinationURL string    `db:"destination_url"`
	Title          string    `db:"title"`
	IsCustomSlug   bool      `db:"is_custom_slug"`
	CreatedAt      time.Time `db:"created_at"`
	UpdatedAt      time.Time `db:"updated_at"`
}

// LinkListItem is a Link plus the click aggregates the list screen needs.
// Kept separate from Link so create/update responses don't carry counters that
// are always zero at that point.
type LinkListItem struct {
	Link
	ClickCount    int64      `db:"click_count"`
	LastClickedAt *time.Time `db:"last_clicked_at"`
}
