package geoip

import (
	"log/slog"
	"net"
	"sync"

	"github.com/oschwald/geoip2-golang"
)

// Resolver wraps MaxMind GeoLite2 database.
// Thread-safe — nhiều goroutines (collector workers) gọi đồng thời OK.
// MaxMind reader dùng memory-mapped file, không cần mutex cho reads.
type Resolver struct {
	db *geoip2.Reader
}

// New mở GeoLite2 database file.
// dbPath là đường dẫn tới file GeoLite2-Country.mmdb
//
// File này ~6MB, loaded vào memory một lần khi app start.
// Lookup sau đó là O(1) — không có network call, không có disk I/O.
func New(dbPath string) (*Resolver, error) {
	db, err := geoip2.Open(dbPath)
	if err != nil {
		return nil, err
	}

	slog.Info("geoip database loaded", "path", dbPath)
	return &Resolver{db: db}, nil
}

// Country trả về ISO country code (VN, US, JP,...) từ IP string.
// Trả về "" nếu không lookup được (private IP, invalid IP, not found).
//
// Tại sao trả "" thay vì error?
// Vì GeoIP lookup fail không phải lỗi nghiêm trọng — analytics data
// thiếu country vẫn tốt hơn là drop cả event.
func (r *Resolver) Country(ip string) string {
	if r == nil || r.db == nil {
		return ""
	}

	parsed := net.ParseIP(ip)
	if parsed == nil {
		return ""
	}

	record, err := r.db.Country(parsed)
	if err != nil {
		return ""
	}

	return record.Country.IsoCode
}

func (r *Resolver) Close() error {
	if r != nil && r.db != nil {
		return r.db.Close()
	}
	return nil
}

// --- Singleton pattern (optional) ---
// Nếu bạn muốn dùng ở nhiều nơi mà không truyền dependency khắp nơi.
// Cá nhân tôi prefer explicit dependency injection qua constructor,
// nhưng singleton cũng hợp lý cho cross-cutting concern như GeoIP.

var (
	globalResolver *Resolver
	once           sync.Once
)

func Init(dbPath string) error {
	var err error
	once.Do(func() {
		globalResolver, err = New(dbPath)
	})
	return err
}

func Global() *Resolver {
	return globalResolver
}
