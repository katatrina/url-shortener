package model

import "time"

type ShortenURLParams struct {
	OriginalURL string
	CustomAlias string // optional: user-chosen short code
	UserID      *string
	ExpiresAt   *time.Time
}

type CreateUserParams struct {
	Email       string
	DisplayName string
	Password    string
}

type LoginUserParams struct {
	Email    string
	Password string
}

type LoginUserResult struct {
	AccessToken          string
	AccessTokenExpiresAt time.Time
	User                 *User
}

type ClickMeta struct {
	IP        string
	UserAgent string
	Referer   string
}

type GetURLStatsParams struct {
	UserID    string
	ShortCode string
	Days      int
}
