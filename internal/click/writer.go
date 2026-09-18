package click

import (
	"context"
	"log/slog"
	"net/netip"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

const insertClicks = `
	INSERT INTO clicks (id, link_id, clicked_at, ip_address, referrer, referrer_host, user_agent, country_code)
	SELECT u.id, u.link_id, u.clicked_at, u.ip, u.referrer, u.referrer_host, u.user_agent, u.country
	FROM unnest($1::uuid[], $2::uuid[], $3::timestamptz[], $4::inet[], $5::text[], $6::text[], $7::text[], $8::text[])
	AS u(id, link_id, clicked_at, ip, referrer, referrer_host, user_agent, country)
	WHERE EXISTS (SELECT 1 FROM links l WHERE l.id = u.link_id)
`

type Writer struct {
	db *pgxpool.Pool
}

func NewWriter(db *pgxpool.Pool) *Writer {
	return &Writer{db: db}
}

func (w *Writer) WriteBatch(ctx context.Context, events []Event) error {
	if len(events) == 0 {
		return nil
	}

	ids := make([]string, len(events))
	linkIDs := make([]string, len(events))
	clickedAts := make([]time.Time, len(events))
	ips := make([]*netip.Addr, len(events))
	referrers := make([]*string, len(events))
	referrerHosts := make([]*string, len(events))
	userAgents := make([]*string, len(events))
	countryCodes := make([]*string, len(events))

	for i := range events {
		e := &events[i]

		ids[i] = e.ID
		linkIDs[i] = e.LinkID
		clickedAts[i] = e.ClickedAt

		if e.IP.IsValid() {
			ips[i] = &e.IP
		}
		if e.Referrer != "" {
			referrers[i] = &e.Referrer
		}
		if e.ReferrerHost != "" {
			referrerHosts[i] = &e.ReferrerHost
		}
		if e.UserAgent != "" {
			userAgents[i] = &e.UserAgent
		}
		if e.CountryCode != "" {
			countryCodes[i] = &e.CountryCode
		}
	}

	tag, err := w.db.Exec(ctx, insertClicks, ids, linkIDs, clickedAts, ips, referrers, referrerHosts, userAgents, countryCodes)
	if err != nil {
		return err
	}

	if skipped := len(events) - int(tag.RowsAffected()); skipped > 0 {
		slog.Warn("click events not inserted",
			slog.Int("count", skipped),
		)
	}

	return nil
}
