package click

import (
	"net/netip"

	"github.com/oschwald/maxminddb-golang/v2"
)

const DefaultGeoIPDBPath = "geoip/dbip-country-lite.mmdb"

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

func (r *MMDBResolver) CountryCode(ip string) string {
	addr, err := netip.ParseAddr(ip)
	if err != nil {
		return ""
	}

	var record struct {
		Country struct {
			ISOCode string `maxminddb:"iso_code"`
		} `maxminddb:"country"`
	}

	if err = r.reader.Lookup(addr).Decode(&record); err != nil {
		return ""
	}

	return record.Country.ISOCode
}

func (r *MMDBResolver) Close() error {
	return r.reader.Close()
}
