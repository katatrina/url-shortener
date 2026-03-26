package analytics

import (
	"context"
	"log"
	"sync"
	"time"
)

// URLStatsRepository defines what the aggregator needs from the persistence layer.
type URLStatsRepository interface {
	AggregateDaily(ctx context.Context, date time.Time) (int64, error)
}

type Aggregator struct {
	repo     URLStatsRepository
	interval time.Duration
	wg       sync.WaitGroup
	doneCh   chan struct{}
}

// NewAggregator creates a new aggregation scheduler.
//
// interval controls how often the job runs. For development, 1 minute is fine
// so you can see results quickly. In production, every 5-15 minutes is typical —
// frequent enough to keep the dashboard fresh, infrequent enough to not
// hammer the database.
func NewAggregator(repo URLStatsRepository, interval time.Duration) *Aggregator {
	return &Aggregator{
		repo:     repo,
		interval: interval,
		doneCh:   make(chan struct{}),
	}
}

// Start launches the aggregation loop in a background goroutine.
func (a *Aggregator) Start() {
	a.wg.Add(1)
	go a.run()
	log.Printf("[INFO] aggregator started: interval=%s", a.interval)
}

// Stop signals the aggregation loop to stop and waits for it to finish.
func (a *Aggregator) Stop() {
	log.Println("[INFO] aggregator stopping...")
	close(a.doneCh)
	a.wg.Wait()
	log.Println("[INFO] aggregator stopped")
}

func (a *Aggregator) run() {
	defer a.wg.Done()

	// Run once immediately on startup so the dashboard has data right away,
	// instead of waiting for the first ticker fire.
	a.aggregate()

	ticker := time.NewTicker(a.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			a.aggregate()
		case <-a.doneCh:
			// Run one final aggregation before shutting down.
			// This ensures any events flushed by the collector during shutdown
			// are captured in the daily stats.
			a.aggregate()
			return
		}
	}
}

func (a *Aggregator) aggregate() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	today := time.Now().UTC().Truncate(24 * time.Hour)

	// Aggregate today's clicks.
	rowsToday, err := a.repo.AggregateDaily(ctx, today)
	if err != nil {
		log.Printf("[ERROR] aggregation failed for %s: %v", today.Format("2006-01-02"), err)
		return
	}

	// Also re-aggregate yesterday, in case late-arriving events were flushed
	// after midnight (e.g., events buffered in the collector at 23:59:58,
	// flushed at 00:00:03). This is a 1-day safety margin.
	yesterday := today.AddDate(0, 0, -1)
	rowsYesterday, err := a.repo.AggregateDaily(ctx, yesterday)
	if err != nil {
		log.Printf("[ERROR] aggregation failed for %s: %v", yesterday.Format("2006-01-02"), err)
		return
	}

	if rowsToday > 0 || rowsYesterday > 0 {
		log.Printf("[DEBUG] aggregation complete: today=%d URLs, yesterday=%d URLs", rowsToday, rowsYesterday)
	}
}
