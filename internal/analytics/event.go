package analytics

import "time"

// ClickEvent represents a single redirect click to be processed asynchronously.
// This struct is deliberately lightweight — it carries just enough data to be
// useful for analytics without bloating the channel buffer memory.
type ClickEvent struct {
	URLID     string
	IP        string
	UserAgent string
	Referer   string
	ClickedAt time.Time
}
