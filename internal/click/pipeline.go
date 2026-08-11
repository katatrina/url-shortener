package click

import (
	"context"
	"log/slog"
	"net/netip"
	"net/url"
	"strings"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
)

const (
	flushTimeout    = 10 * time.Second
	maxReferrerLen  = 2048
	maxUserAgentLen = 512
)

type Event struct {
	ID        string
	LinkID    string
	ClickedAt time.Time

	IP           netip.Addr
	Referrer     string
	ReferrerHost string
	UserAgent    string
	CountryCode  string
}

func NewEvent(linkID string, ip string, referrer string, userAgent string) Event {
	id, _ := uuid.NewV7()
	addr, _ := netip.ParseAddr(ip)
	addr = addr.Unmap()

	return Event{
		ID:        id.String(),
		LinkID:    linkID,
		ClickedAt: time.Now(),
		IP:        addr,
		Referrer:  referrer,
		UserAgent: userAgent,
	}
}

type BatchWriter interface {
	WriteBatch(ctx context.Context, events []Event) error
}

type CountryResolver interface {
	CountryCode(ip netip.Addr) string
}

type Pipeline struct {
	events      chan Event
	writer      BatchWriter
	resolver    CountryResolver
	batchSize   int
	interval    time.Duration
	dropped     atomic.Uint64
	writeFailed atomic.Uint64
}

func NewPipeline(w BatchWriter, r CountryResolver, bufferSize, batchSize int, interval time.Duration) *Pipeline {
	return &Pipeline{
		events:    make(chan Event, bufferSize),
		writer:    w,
		resolver:  r,
		batchSize: batchSize,
		interval:  interval,
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
			e.Referrer = e.Referrer[:maxReferrerLen]
		}
		e.Referrer = strings.ToValidUTF8(e.Referrer, "")
		e.ReferrerHost = referrerHost(e.Referrer)

		if len(e.UserAgent) > maxUserAgentLen {
			e.UserAgent = e.UserAgent[:maxUserAgentLen]
		}
		e.UserAgent = strings.ToValidUTF8(e.UserAgent, "")

		e.CountryCode = p.resolver.CountryCode(e.IP)
	}

	ctx, cancel := context.WithTimeout(context.Background(), flushTimeout)
	defer cancel()

	if err := p.writer.WriteBatch(ctx, events); err != nil {
		p.writeFailed.Add(uint64(len(events)))
		slog.Error("click batch write failed",
			slog.Int("count", len(events)),
			slog.Uint64("failed_total", p.writeFailed.Load()),
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

func referrerHost(referrer string) string {
	if referrer == "" {
		return ""
	}

	u, err := url.Parse(referrer)
	if err != nil {
		return ""
	}

	host := strings.ToLower(u.Hostname())
	return strings.TrimPrefix(host, "www.")
}
