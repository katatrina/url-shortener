package click

import (
	"context"
	"log/slog"
	"strings"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
)

const (
	DefaultBufferSize     = 1024
	DefaultFlushBatchSize = 100
	DefaultFlushInterval  = 5 * time.Second

	flushTimeout   = 10 * time.Second
	maxReferrerLen = 2048
)

type Event struct {
	ID        string
	LinkID    string
	ClickedAt time.Time

	IP          string
	Referrer    string
	CountryCode string
}

func NewEvent(linkID string, ip string, referrer string) Event {
	id, _ := uuid.NewV7()

	return Event{
		ID:        id.String(),
		LinkID:    linkID,
		ClickedAt: time.Now(),
		IP:        ip,
		Referrer:  referrer,
	}
}

type BatchWriter interface {
	WriteBatch(ctx context.Context, events []Event) error
}

type CountryResolver interface {
	CountryCode(ip string) string
}

type Pipeline struct {
	events          chan Event
	writer          BatchWriter
	countryResolver CountryResolver
	batchSize       int
	interval        time.Duration
	dropped         atomic.Uint64
}

func NewPipeline(w BatchWriter, cr CountryResolver, bufferSize, batchSize int, interval time.Duration) *Pipeline {
	return &Pipeline{
		events:          make(chan Event, bufferSize),
		writer:          w,
		countryResolver: cr,
		batchSize:       batchSize,
		interval:        interval,
	}
}

func (p *Pipeline) Record(e Event) {
	select {
	case p.events <- e:
	default:
		p.dropped.Add(1)
		slog.Warn("click event dropped: buffer full",
			slog.String("link_id", e.LinkID),
			slog.Uint64("dropped_total", p.dropped.Load()),
		)
	}
}

func (p *Pipeline) Dropped() uint64 {
	return p.dropped.Load()
}

func (p *Pipeline) Run(ctx context.Context) {
	ticker := time.NewTicker(p.interval)
	defer ticker.Stop()

	batch := make([]Event, 0, p.batchSize)

	for {
		select {
		case e := <-p.events:
			batch = append(batch, e)
			if len(batch) >= p.batchSize {
				p.flush(&batch)
			}
		case <-ticker.C:
			if len(batch) > 0 {
				p.flush(&batch)
			}
		case <-ctx.Done():
			p.drain(&batch)
			return
		}
	}
}

func (p *Pipeline) flush(batch *[]Event) {
	events := *batch
	for i := range events {
		e := &events[i]

		if len(e.Referrer) > maxReferrerLen {
			e.Referrer = strings.ToValidUTF8(e.Referrer[:maxReferrerLen], "")
		}

		e.CountryCode = p.countryResolver.CountryCode(e.IP)
	}

	ctx, cancel := context.WithTimeout(context.Background(), flushTimeout)
	defer cancel()

	if err := p.writer.WriteBatch(ctx, events); err != nil {
		slog.Error("click batch write failed",
			slog.Int("count", len(events)),
			slog.Any("error", err))
	}

	*batch = (*batch)[:0]
}

func (p *Pipeline) drain(batch *[]Event) {
	for {
		select {
		case e := <-p.events:
			*batch = append(*batch, e)
			if len(*batch) >= p.batchSize {
				p.flush(batch)
			}
		default:
			if len(*batch) > 0 {
				p.flush(batch)
			}
			return
		}
	}
}
