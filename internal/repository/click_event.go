package repository

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/katatrina/url-shortener/internal/model"
)

type ClickEventRepository struct {
	db *pgxpool.Pool
}

func NewClickEventRepository(db *pgxpool.Pool) *ClickEventRepository {
	return &ClickEventRepository{db}
}

// BulkInsert inserts multiple click events in a single database round trip
// using PostgreSQL's COPY protocol (pgx.CopyFrom).
//
// Why COPY instead of multi-row INSERT?
// - INSERT with 100 rows: Postgres parses SQL, plans query, executes.
// - COPY with 100 rows: Postgres uses a streaming binary protocol,
//   skipping SQL parsing entirely. For raw bulk inserts, COPY is the fastest.
//
// For context: pgx's CopyFrom maps to Postgres's "COPY ... FROM STDIN" command.
func (r *ClickEventRepository) BulkInsert(ctx context.Context, events []model.ClickEvent) error {
	rows := make([][]any, len(events))

	for i, e := range events {
		rows[i] = []any{
			e.ID,
			e.URLID,
			e.IP,
			e.Referer,
			e.UserAgentRaw,
			e.OS,
			e.Browser,
			e.DeviceType,
			e.Country,
			e.ClickedAt,
		}
	}

	// CopyFrom works like all-or-nothing (similar to transaction)
	_, err := r.db.CopyFrom(
		ctx,
		pgx.Identifier{"click_events"},
		[]string{"id", "url_id", "ip_address", "referer", "user_agent_raw", "os", "browser", "device_type", "country", "clicked_at"},
		pgx.CopyFromRows(rows),
	)

	return err
}
