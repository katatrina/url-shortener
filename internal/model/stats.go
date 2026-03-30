package model

import "time"

type DailyStat struct {
	Date       time.Time `db:"date" json:"date"`
	ClickCount int64     `db:"click_count" json:"clickCount"`
}

type DailyStatResponse struct {
	Date       int64 `json:"date"`
	ClickCount int64 `json:"clickCount"`
}

type ReferrerStat struct {
	Referer    string `db:"referer" json:"referer"`
	ClickCount int64  `db:"click_count" json:"clickCount"`
}

type CountryStat struct {
	Country    string `db:"country" json:"country"`
	ClickCount int64  `db:"click_count" json:"clickCount"`
}

type URLStatsResponse struct {
	TotalClicks  int64               `json:"totalClicks"`
	DailyClicks  []DailyStatResponse `json:"dailyClicks"`
	TopReferrers []ReferrerStat      `json:"topReferrers"`
	TopCountries []CountryStat       `json:"topCountries"`
}
