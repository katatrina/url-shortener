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

type OSStat struct {
	OS         string `db:"os" json:"os"`
	ClickCount int64  `db:"click_count" json:"clickCount"`
}

type BrowserStat struct {
	Browser    string `db:"browser" json:"browser"`
	ClickCount int64  `db:"click_count" json:"clickCount"`
}

type DeviceStat struct {
	DeviceType string `db:"device_type" json:"deviceType"`
	ClickCount int64  `db:"click_count" json:"clickCount"`
}

type URLStatsResponse struct {
	TotalClicks  int64               `json:"totalClicks"`
	DailyClicks  []DailyStatResponse `json:"dailyClicks"`
	TopReferrers []ReferrerStat      `json:"topReferrers"`
	TopCountries []CountryStat       `json:"topCountries"`
	TopOSes      []OSStat            `json:"topOSes"`
	TopBrowsers  []BrowserStat       `json:"topBrowsers"`
	TopDevices   []DeviceStat        `json:"topDevices"`
}
