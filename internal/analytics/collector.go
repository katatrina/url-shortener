package analytics

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/katatrina/url-shortener/internal/model"
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

type ClickCollector struct {
	eventCh     chan model.ClickEvent
	eventWriter ClickEventWriter
	cfg         ClickCollectorConfig
	wg          sync.WaitGroup
}

func NewClickCollector(eventWriter ClickEventWriter, cfg ClickCollectorConfig) *ClickCollector {
	return &ClickCollector{
		eventCh:     make(chan model.ClickEvent, cfg.ChannelBuffer),
		eventWriter: eventWriter,
		cfg:         cfg,
	}
}

// Start launches the worker pool. Call this once at application startup.
// Each worker runs in its own goroutine, reading from the shared event channel.
func (c *ClickCollector) Start() {
	for i := range c.cfg.WorkerCount {
		c.wg.Add(1)
		go c.worker(i)
	}
	log.Printf("[INFO] analytics collector started: %d workers, buffer=%d, batch=%d, flush=%s",
		c.cfg.WorkerCount, c.cfg.ChannelBuffer, c.cfg.BatchSize, c.cfg.FlushInterval)
}

// Stop signals all workers to shut down and waits for them to finish.
//
// How graceful shutdown works:
// 1. close(c.eventCh) tells workers "no more events are coming"
// 2. Workers finish processing any events still in the channel buffer
// 3. Workers flush their final partial batch
// 4. c.wg.Wait() blocks until every worker has exited
//
// After Stop returns, it's guaranteed that all events have been persisted.
func (c *ClickCollector) Stop() {
	log.Println("[INFO] analytics collector stopping, draining remaining events...")
	close(c.eventCh)
	c.wg.Wait()
	log.Println("[INFO] analytics collector stopped")
}

// Track queues a click event for async processing.
// This is called from the redirect handler on every click.
//
// Non-blocking: if the channel is full, the event is dropped rather than
// blocking the HTTP response. This is a deliberate trade-off:
//   - User experience > analytics accuracy
//   - A dropped analytics event is invisible to the user
//   - A blocked redirect is a terrible user experience
//
// In practice, with a 10,000 buffer and workers draining continuously,
// drops should be extremely rare. If you're seeing drops in logs,
// it means you need more workers or a bigger buffer.
func (c *ClickCollector) Track(urlID string, meta model.ClickMeta) {
	id, _ := uuid.NewV7()
	event := model.ClickEvent{
		ID:        id.String(),
		URLID:     urlID,
		IP:        toNullable(meta.IP),
		UserAgent: toNullable(meta.UserAgent),
		Referer:   toNullable(meta.Referer),
		ClickedAt: time.Now(),
	}

	select {
	case c.eventCh <- event:
	default:
		log.Printf("[WARN] channel full, dropping event for url=%s", urlID)
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
			log.Printf("[ERROR] worker %d: batch insert failed (%d events lost): %v", id, len(batch), err)
		} else {
			log.Printf("[DEBUG] worker %d: flushed %d events", id, len(batch))
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
				log.Printf("[INFO] worker %d: shutdown complete", id)
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
		}
	}
}
