package analytics

import (
	"context"
	"log/slog"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"github.com/katatrina/url-shortener/internal/metrics"
	"github.com/katatrina/url-shortener/internal/model"
	ua "github.com/mileusna/useragent"
)

type ClickEventWriter interface {
	BulkInsert(ctx context.Context, events []model.ClickEvent) error
}

// ClickCollectorConfig holds tuning parameters for the analytics pipeline.
// These are separated from ClickCollector to make them configurable (e.g., from env vars).
type ClickCollectorConfig struct {
	WorkerCount   int           // Number of goroutines consuming events
	ChannelBuffer int           // Max events queued before backpressure kicks in
	BatchSize     int           // Flush to DB when this many events accumulate
	FlushInterval time.Duration // Flush to DB at least this often, even if batch isn't full
}

// DefaultCollectorConfig returns sensible defaults for development.
// In production, these would be tuned based on traffic volume and DB capacity.
func DefaultCollectorConfig() ClickCollectorConfig {
	return ClickCollectorConfig{
		WorkerCount:   2,               // 2 workers is enough for moderate traffic
		ChannelBuffer: 10000,           // Buffer up to 10k events before dropping
		BatchSize:     100,             // Insert 100 rows per query
		FlushInterval: 5 * time.Second, // Don't hold events longer than 5s
	}
}

type GeoResolver interface {
	Country(ip string) string
}

type ClickCollector struct {
	eventCh     chan model.ClickEvent
	eventWriter ClickEventWriter
	cfg         ClickCollectorConfig
	geoResolver GeoResolver
	wg          sync.WaitGroup
	stopped     atomic.Bool
}

func NewClickCollector(
	eventWriter ClickEventWriter,
	cfg ClickCollectorConfig,
	geoResolver GeoResolver,
) *ClickCollector {
	return &ClickCollector{
		eventCh:     make(chan model.ClickEvent, cfg.ChannelBuffer),
		eventWriter: eventWriter,
		cfg:         cfg,
		geoResolver: geoResolver,
	}
}

// Start launches the worker pool. Call this once at application startup.
// Each worker runs in its own goroutine, reading from the shared event channel.
func (c *ClickCollector) Start() {
	for i := range c.cfg.WorkerCount {
		c.wg.Add(1)
		go c.worker(i)
	}
	slog.Info("analytics collector started",
		"workers", c.cfg.WorkerCount, "buffer", c.cfg.ChannelBuffer,
		"batch_size", c.cfg.BatchSize, "flush_interval", c.cfg.FlushInterval)
}

// Stop signals all workers to shut down and waits for them to finish.
//
// How graceful shutdown works:
// 1. c.stopped is set to true — Track() stops accepting new events
// 2. close(c.eventCh) tells workers "no more events are coming"
// 3. Workers finish processing any events still in the channel buffer
// 4. Workers flush their final partial batch
// 5. c.wg.Wait() blocks until every worker has exited
//
// Why stopped flag before close?
// Track() is called via `go` from the HTTP handler. During shutdown, there's a
// race window between srv.Shutdown() returning and close(eventCh): a goroutine
// spawned just before shutdown could try to send on a closed channel → panic.
// The atomic flag ensures Track() silently drops events once shutdown begins,
// making the close safe.
//
// After Stop returns, it's guaranteed that all buffered events have been persisted.
func (c *ClickCollector) Stop() {
	slog.Info("analytics collector stopping, draining remaining events...")
	c.stopped.Store(true)
	close(c.eventCh)
	c.wg.Wait()
	slog.Info("analytics collector stopped")
}

// Track queues a click event for async processing.
// This is called from the Redirect service method on every click.
//
// Non-blocking: if the channel is full or the collector is shutting down,
// the event is dropped rather than blocking the HTTP response.
// This is a deliberate trade-off:
//   - User experience > analytics accuracy
//   - A dropped analytics event is invisible to the user
//   - A blocked redirect is a terrible user experience
//
// In practice, with a 10,000 buffer and workers draining continuously,
// drops should be extremely rare. If you're seeing drops in logs,
// it means you need more workers or a bigger buffer.
func (c *ClickCollector) Track(urlID string, meta model.ClickMeta) {
	// Check shutdown flag first to avoid sending on a closed channel.
	// This is the only guard needed: once stopped is true, Stop() will
	// close the channel. Without this check, a late goroutine could panic.
	if c.stopped.Load() {
		return
	}

	id, _ := uuid.NewV7()
	country := c.geoResolver.Country(meta.IP) // Lookup from memory-mapped file, ~1 microsecond per call

	// Parse User-Agent string to extract OS, browser, and device type.
	// Pure CPU operation (string parsing), no I/O — safe to call inline.
	parsed := ua.Parse(meta.UserAgent)

	event := model.ClickEvent{
		ID:           id.String(),
		URLID:        urlID,
		IP:           toNullable(meta.IP),
		Referer:      toNullable(meta.Referer),
		UserAgentRaw: toNullable(meta.UserAgent),
		OS:           toNullable(parsed.OS),
		Browser:      toNullable(parsed.Name),
		DeviceType:   toNullable(detectDevice(parsed)),
		Country:      toNullable(country),
		ClickedAt:    time.Now(),
	}

	select {
	case c.eventCh <- event:
	default:
		// Channel full — event dropped.
		metrics.AnalyticsEventsDropped.Inc()
		slog.Warn("channel full, dropping event", "url_id", urlID)
	}
}

func detectDevice(parsed ua.UserAgent) string {
	switch {
	case parsed.Bot:
		return "Bot"
	case parsed.Tablet:
		return "Tablet"
	case parsed.Mobile:
		return "Mobile"
	case parsed.Desktop:
		return "Desktop"
	default:
		return ""
	}
}

func toNullable(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// QueueDepth returns the number of events waiting to be processed.
// Useful for metrics/monitoring (Phase 4: Prometheus gauge).
func (c *ClickCollector) QueueDepth() int {
	return len(c.eventCh)
}

// worker is the main loop for each worker goroutine.
// It accumulates events into a batch and flushes when either:
//   - The batch reaches BatchSize (flush-on-count)
//   - The FlushInterval timer fires (flush-on-time)
//   - The channel is closed (graceful shutdown)
//
// The "flush on count OR time, whichever comes first" pattern is classic
// in stream processing. Without the timer, low-traffic periods would cause
// events to sit in memory indefinitely. Without the count threshold,
// high-traffic periods would wait for the timer even when the batch is full.
func (c *ClickCollector) worker(id int) {
	defer c.wg.Done()

	batch := make([]model.ClickEvent, 0, c.cfg.BatchSize)
	ticker := time.NewTicker(c.cfg.FlushInterval)
	defer ticker.Stop()

	flush := func() {
		if len(batch) == 0 {
			return
		}

		if err := c.eventWriter.BulkInsert(context.Background(), batch); err != nil {
			// DB insert failed — log and discard the batch.
			// In a production system, you might push failed events to a dead letter queue
			// or retry with backoff. For now, logging is sufficient.
			metrics.AnalyticsBatchErrors.Inc()
			slog.Error("batch insert failed", "worker", id, "events_lost", len(batch), "error", err)
		} else {
			metrics.AnalyticsEventsInserted.Add(float64(len(batch)))
			metrics.AnalyticsBatchFlushTotal.Inc()
			slog.Debug("flushed events", "worker", id, "count", len(batch))
		}

		// Reset the batch. We keep the underlying array to avoid re-allocation.
		batch = batch[:0]
	}

	for {
		select {
		case event, ok := <-c.eventCh:
			if !ok {
				// Channel closed — this is graceful shutdown.
				// Flush whatever we have and exit.
				flush()
				slog.Info("worker shutdown complete", "worker", id)
				return
			}

			batch = append(batch, event)
			if len(batch) >= c.cfg.BatchSize {
				flush()
				// Reset the ticker after a count-based flush.
				// Without this, the ticker might fire immediately after a full-batch flush,
				// causing an unnecessary empty flush.
				ticker.Reset(c.cfg.FlushInterval)
			}

		case <-ticker.C:
			// Timer fired — flush partial batch to avoid holding events too long.
			flush()
			metrics.AnalyticsQueueDepth.Set(float64(len(c.eventCh)))
		}
	}
}
