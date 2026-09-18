package click

import (
	"log/slog"
	"net/netip"

	"github.com/oschwald/maxminddb-golang/v2"
)

type NoopResolver struct{}

func (NoopResolver) CountryCode(netip.Addr) string { return "" }

type MMDBResolver struct {
	reader *maxminddb.Reader
}

func NewMMDBResolver(path string) (*MMDBResolver, error) {
	reader, err := maxminddb.Open(path)
	if err != nil {
		return nil, err
	}

	return &MMDBResolver{reader: reader}, nil
}

func (r *MMDBResolver) CountryCode(addr netip.Addr) string {
	var record struct {
		Country struct {
			ISOCode string `maxminddb:"iso_code"`
		} `maxminddb:"country"`
	}

	if err := r.reader.Lookup(addr).Decode(&record); err != nil {
		slog.Warn("country lookup failed", slog.Any("error", err))
		return ""
	}

	return record.Country.ISOCode
}

func (r *MMDBResolver) Close() error {
	return r.reader.Close()
}
