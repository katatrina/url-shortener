package model

import "time"

type ClickEvent struct {
	ID        string
	URLID     string
	IP        *string
	UserAgent *string
	Referer   *string
	Country   *string
	ClickedAt time.Time
}
