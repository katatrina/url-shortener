package model

import "time"

type ClickEvent struct {
	ID           string
	URLID        string
	IP           *string
	Referer      *string
	UserAgentRaw *string
	OS           *string
	Browser      *string
	DeviceType   *string
	Country      *string
	ClickedAt    time.Time
}
