package metrics

import (
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prometheus/client_golang/prometheus"
)

// DBPoolCollector adapts pgxpool.Stat() to Prometheus metrics.
//
// Why a custom collector instead of simple Gauges?
//
// With simple Gauges, you'd need a goroutine that periodically calls
// pool.Stat() and updates each gauge. That's extra complexity: another
// goroutine to manage, another shutdown sequence to worry about.
//
// A custom Collector is lazy — Prometheus calls Collect() only when it
// scrapes /metrics. No background goroutine needed. The pool stats are
// read fresh at scrape time, which is exactly what we want.
//
// This is the idiomatic Prometheus way to expose "live" state from
// an external system (connection pool, thread pool, etc).
type DBPoolCollector struct {
	pool *pgxpool.Pool

	// Descriptors describe each metric. Prometheus needs these upfront
	// (via Describe) to validate that Collect produces consistent metrics.
	// Think of them as "metadata" — name, help text, labels.
	acquireCount         *prometheus.Desc
	acquireDuration      *prometheus.Desc
	acquiredConns        *prometheus.Desc
	idleConns            *prometheus.Desc
	totalConns           *prometheus.Desc
	maxConns             *prometheus.Desc
	emptyAcquireCount    *prometheus.Desc
	canceledAcquireCount *prometheus.Desc
}

func NewDBPoolCollector(pool *pgxpool.Pool) *DBPoolCollector {
	ns := namespace
	sub := "db_pool"

	return &DBPoolCollector{
		pool: pool,

		// acquireCount: cumulative number of times a connection was acquired
		// from the pool. This is a Counter (always increases).
		//
		// High acquire rate + high acquireDuration = pool is a bottleneck.
		acquireCount: prometheus.NewDesc(
			prometheus.BuildFQName(ns, sub, "acquire_count_total"),
			"Total number of times a connection was acquired from the pool.",
			nil, nil,
		),

		// acquireDuration: cumulative time spent waiting for a connection.
		// If this grows much faster than acquireCount, each acquire is
		// taking longer → pool exhaustion is approaching.
		//
		// Average wait time = acquireDuration / acquireCount
		acquireDuration: prometheus.NewDesc(
			prometheus.BuildFQName(ns, sub, "acquire_duration_seconds_total"),
			"Total time spent waiting to acquire a connection.",
			nil, nil,
		),

		// acquiredConns: connections currently checked out (in use by queries).
		// This is a Gauge (goes up and down).
		//
		// If acquiredConns == maxConns for sustained periods, the pool is
		// saturated and new queries must wait.
		acquiredConns: prometheus.NewDesc(
			prometheus.BuildFQName(ns, sub, "acquired_conns"),
			"Number of connections currently acquired (in use).",
			nil, nil,
		),

		// idleConns: connections sitting in the pool, waiting to be used.
		// Healthy pool: some idle conns always available.
		// Unhealthy: 0 idle for extended periods (all connections busy).
		idleConns: prometheus.NewDesc(
			prometheus.BuildFQName(ns, sub, "idle_conns"),
			"Number of idle connections in the pool.",
			nil, nil,
		),

		// totalConns: idle + acquired. Should be <= maxConns.
		// If totalConns < maxConns, pool hasn't needed to open all
		// connections yet (traffic is low). Good.
		totalConns: prometheus.NewDesc(
			prometheus.BuildFQName(ns, sub, "total_conns"),
			"Total number of connections (idle + acquired).",
			nil, nil,
		),

		// maxConns: maximum pool size (from pgxpool config).
		// This is the ceiling. Useful as a reference line on dashboards.
		// pgxpool defaults to max(4, runtime.NumCPU()) connections.
		maxConns: prometheus.NewDesc(
			prometheus.BuildFQName(ns, sub, "max_conns"),
			"Maximum number of connections allowed in the pool.",
			nil, nil,
		),

		// emptyAcquireCount: how many times a caller had to WAIT because
		// no idle connection was available. This is the early warning sign
		// of pool pressure — it's not an error yet, but if it keeps climbing,
		// saturation is coming.
		emptyAcquireCount: prometheus.NewDesc(
			prometheus.BuildFQName(ns, sub, "empty_acquire_count_total"),
			"Total acquires that had to wait for a connection (pool was empty).",
			nil, nil,
		),

		// canceledAcquireCount: acquires that were canceled before getting
		// a connection (usually because the request's context timed out
		// while waiting). This IS an error from the user's perspective —
		// their request failed because they couldn't get a DB connection.
		canceledAcquireCount: prometheus.NewDesc(
			prometheus.BuildFQName(ns, sub, "canceled_acquire_count_total"),
			"Total acquires canceled while waiting (context deadline exceeded).",
			nil, nil,
		),
	}
}

// Describe sends metric descriptors to Prometheus.
// Called once at registration time.
func (c *DBPoolCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.acquireCount
	ch <- c.acquireDuration
	ch <- c.acquiredConns
	ch <- c.idleConns
	ch <- c.totalConns
	ch <- c.maxConns
	ch <- c.emptyAcquireCount
	ch <- c.canceledAcquireCount
}

// Collect reads current pool stats and sends them to Prometheus.
// Called every time Prometheus scrapes /metrics.
func (c *DBPoolCollector) Collect(ch chan<- prometheus.Metric) {
	stat := c.pool.Stat()

	ch <- prometheus.MustNewConstMetric(c.acquireCount, prometheus.CounterValue,
		float64(stat.AcquireCount()))

	ch <- prometheus.MustNewConstMetric(c.acquireDuration, prometheus.CounterValue,
		stat.AcquireDuration().Seconds())

	ch <- prometheus.MustNewConstMetric(c.acquiredConns, prometheus.GaugeValue,
		float64(stat.AcquiredConns()))

	ch <- prometheus.MustNewConstMetric(c.idleConns, prometheus.GaugeValue,
		float64(stat.IdleConns()))

	ch <- prometheus.MustNewConstMetric(c.totalConns, prometheus.GaugeValue,
		float64(stat.TotalConns()))

	ch <- prometheus.MustNewConstMetric(c.maxConns, prometheus.GaugeValue,
		float64(stat.MaxConns()))

	ch <- prometheus.MustNewConstMetric(c.emptyAcquireCount, prometheus.CounterValue,
		float64(stat.EmptyAcquireCount()))

	ch <- prometheus.MustNewConstMetric(c.canceledAcquireCount, prometheus.CounterValue,
		float64(stat.CanceledAcquireCount()))
}
