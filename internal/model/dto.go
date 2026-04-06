package model

import "time"

type ShortenURLParams struct {
	LongURL     string
	CustomAlias string
	UserID      *string
	ExpiresAt   *time.Time
}

type CreateUserParams struct {
	Email    string
	FullName string
	Password string
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
