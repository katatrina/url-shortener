package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/katatrina/url-shortener/internal/analytics"
)

type ClickEventRepository struct {
	db *pgxpool.Pool
}

func NewClickEventRepository(db *pgxpool.Pool) *ClickEventRepository {
	return &ClickEventRepository{db}
}

// BulkInsert inserts multiple click events in a single database round-trip
// using PostgreSQL's COPY protocol (pgx.CopyFrom).
//
// Why COPY instead of multi-row INSERT?
// - INSERT with 100 rows: Postgres parses SQL, plans query, executes.
// - COPY with 100 rows: Postgres uses a streaming binary protocol,
//   skipping SQL parsing entirely. For bulk inserts, COPY is 2-5x faster.
//
// For context: pgx's CopyFrom maps to Postgres's "COPY ... FROM STDIN" command,
// which is the standard way to bulk-load data in Postgres.
func (r *ClickEventRepository) BulkInsert(ctx context.Context, events []analytics.ClickEvent) error {
	rows := make([][]any, len(events))

	for i, e := range events {
		id, err := uuid.NewV7()
		if err != nil {
			return fmt.Errorf("failed to generate click event ID: %w", err)
		}

		rows[i] = []any{
			id.String(),
			e.URLID,
			e.IP,
			e.UserAgent,
			e.Referer,
			nil, // country — will be populated by geo-IP enrichment later
			e.ClickedAt,
		}
	}

	copied, err := r.db.CopyFrom(
		ctx,
		pgx.Identifier{"click_events"},
		[]string{"id", "url_id", "ip_address", "user_agent", "referer", "country", "clicked_at"},
		pgx.CopyFromRows(rows),
	)

	if err != nil {
		return fmt.Errorf("batch insert failed: %w", err)
	}

	if copied != int64(len(events)) {
		return fmt.Errorf("expected to insert %d rows, but inserted %d", len(events), copied)
	}

	return nil
}
