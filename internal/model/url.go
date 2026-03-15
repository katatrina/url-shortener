package model

import "time"

type URL struct {
	ID          string     `db:"id"`
	ShortCode   string     `db:"short_code"`
	OriginalURL string     `db:"original_url"`
	UserID      *string    `db:"user_id"`
	ClickCount  int64      `db:"click_count"`
	ExpiresAt   *time.Time `db:"expires_at"`
	CreatedAt   time.Time  `db:"created_at"`
	UpdatedAt   time.Time  `db:"updated_at"`
	DeletedAt   *time.Time `db:"deleted_at"`
}

// IsExpired returns true if the URL has an expiry date that has passed.
func (u *URL) IsExpired() bool {
	if u.ExpiresAt == nil {
		return false
	}

	return time.Now().After(*u.ExpiresAt)
}
