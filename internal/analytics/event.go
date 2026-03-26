package analytics

import (
	"time"

	"github.com/google/uuid"
)

type ClickEvent struct {
	ID        string
	URLID     string
	IP        *string
	UserAgent *string
	Referer   *string
	Country   *string
	ClickedAt time.Time
}

func NewClickEvent(urlID, ip, userAgent, referer string) ClickEvent {
	id, _ := uuid.NewV7()

	return ClickEvent{
		ID:        id.String(),
		URLID:     urlID,
		IP:        toNullable(ip),
		UserAgent: toNullable(userAgent),
		Referer:   toNullable(referer),
		ClickedAt: time.Now(),
	}
}

func toNullable(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
