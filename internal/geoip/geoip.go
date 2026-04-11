package geoip

import (
	"net/netip"

	"github.com/oschwald/geoip2-golang/v2"
)

// GeoService wraps MaxMind reader.
// Dùng struct để dễ mock trong unit tests.
type GeoService struct {
	db *geoip2.Reader
}

// New mở file .mmdb và load vào memory.
// Gọi hàm này MỘT LẦN DUY NHẤT lúc app khởi động - không bao giờ gọi per-request.
func New(dbPath string) (*GeoService, error) {
	db, err := geoip2.Open(dbPath)
	if err != nil {
		return nil, err
	}
	return &GeoService{db}, nil
}

// Close giải phóng file handle.
// Gọi khi app shutdown - dùng defer trong run().
func (g *GeoService) Close() error {
	return g.db.Close()
}

// CountryCode nhận IP string, trả về ISO 3166-1 alpha-2 country code (VD: "VN", "US").
func (g *GeoService) CountryCode(ipStr string) string {
	ip, err := netip.ParseAddr(ipStr)
	if err != nil {
		return "" // IP string malformed
	}

	record, err := g.db.Country(ip)
	if err != nil {
		return "" // IP hợp lệ nhưng không có trong database
	}

	if !record.HasData() {
		return "" // Private IP, reserved range
	}

	if record.Country.ISOCode == "" {
		return ""
	}

	return record.Country.ISOCode
}
