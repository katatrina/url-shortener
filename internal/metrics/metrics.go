package metrics

import "github.com/prometheus/client_golang/prometheus"

// -------------------------------------------------------------------
// Metric naming convention trong Prometheus:
//   <namespace>_<subsystem>_<name>_<unit>
//
// Ví dụ: urlshortener_http_requests_total
//   - namespace:  urlshortener  (tên app)
//   - subsystem:  http          (component nào)
//   - name:       requests      (đo cái gì)
//   - unit/suffix: _total       (counter thì thêm _total)
//
// Histogram đo thời gian thì suffix là _seconds (không phải _ms).
// Prometheus convention dùng base unit: seconds, bytes, meters...
// -------------------------------------------------------------------

const namespace = "urlshortener"

// ========================
// HTTP Metrics
// ========================

// HTTPRequestsTotal counts every HTTP request that hits the server.
//
// Labels allow us to "slice" this counter by different dimensions:
//   - method: GET, POST, DELETE...
//   - path:   the route pattern (e.g. "/api/v1/shorten", not the actual URL with params)
//   - status: HTTP status code as string ("200", "404", "500")
//
// Why these 3 labels? Because they answer the most common debugging questions:
//   "Which endpoint is getting hammered?" → filter by path
//   "Are we returning errors?"           → filter by status
//   "Is it reads or writes?"             → filter by method
//
// WARNING: Don't use high-cardinality labels like user_id or full URL path.
// Each unique label combination creates a new time series in Prometheus.
// 1000 users × 10 endpoints × 5 status codes = 50,000 time series.
// That can kill Prometheus's memory. Keep labels low-cardinality.
var HTTPRequestsTotal = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Namespace: namespace,
		Subsystem: "http",
		Name:      "requests_total",
		Help:      "Total number of HTTP requests processed.",
	},
	[]string{"method", "path", "status"},
)

// HTTPRequestDuration measures how long each HTTP request takes.
//
// We use a Histogram because we care about the DISTRIBUTION of latency,
// not just the average. An average of 50ms could mean "all requests ~50ms"
// or "99% at 5ms, 1% at 5 seconds" — very different situations.
//
// The Buckets define the "boundaries" for counting requests:
//   - 5ms, 10ms, 25ms, 50ms, 100ms, 250ms, 500ms, 1s, 2.5s, 5s
//
// These are chosen based on typical web app latency ranges:
//   - < 50ms: excellent (cache hit)
//   - 50-100ms: good (simple DB query)
//   - 100-500ms: acceptable (complex query)
//   - > 500ms: needs investigation
//   - > 1s: something is wrong
//
// You can adjust these based on your SLA. For a URL shortener,
// the redirect endpoint should be < 50ms for cache hits.
var HTTPRequestDuration = prometheus.NewHistogramVec(
	prometheus.HistogramOpts{
		Namespace: namespace,
		Subsystem: "http",
		Name:      "request_duration_seconds",
		Help:      "Duration of HTTP requests in seconds.",
		Buckets:   []float64{0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5},
	},
	[]string{"method", "path"},
)

// HTTPRequestsInFlight tracks how many requests are being handled RIGHT NOW.
// This is a Gauge because it goes up (request starts) and down (request ends).
//
// Why is this useful? If this number keeps climbing and never comes down,
// you have a goroutine leak or requests are getting stuck somewhere.
// It's also a good indicator of server load at any given moment.
var HTTPRequestsInFlight = prometheus.NewGauge(
	prometheus.GaugeOpts{
		Namespace: namespace,
		Subsystem: "http",
		Name:      "requests_in_flight",
		Help:      "Number of HTTP requests currently being processed.",
	},
)

// ========================
// Analytics Pipeline Metrics
// ========================

// AnalyticsQueueDepth shows how many click events are waiting in the channel.
// If this number stays high, workers can't keep up with incoming events.
// If it hits the channel buffer size (10,000), events start dropping.
var AnalyticsQueueDepth = prometheus.NewGauge(
	prometheus.GaugeOpts{
		Namespace: namespace,
		Subsystem: "analytics",
		Name:      "queue_depth",
		Help:      "Number of click events waiting in the collector channel.",
	},
)

// AnalyticsEventsDropped counts events that were dropped because the channel was full.
// In a healthy system this should always be 0.
// If you see this increasing, you need more workers or a bigger buffer.
var AnalyticsEventsDropped = prometheus.NewCounter(
	prometheus.CounterOpts{
		Namespace: namespace,
		Subsystem: "analytics",
		Name:      "events_dropped_total",
		Help:      "Total number of click events dropped due to full channel.",
	},
)

// AnalyticsBatchFlushTotal counts how many times workers flushed a batch to DB.
// Combined with AnalyticsEventsInserted, you can calculate average batch size.
var AnalyticsBatchFlushTotal = prometheus.NewCounter(
	prometheus.CounterOpts{
		Namespace: namespace,
		Subsystem: "analytics",
		Name:      "batch_flush_total",
		Help:      "Total number of batch flushes to database.",
	},
)

// AnalyticsEventsInserted counts the total events successfully written to DB.
var AnalyticsEventsInserted = prometheus.NewCounter(
	prometheus.CounterOpts{
		Namespace: namespace,
		Subsystem: "analytics",
		Name:      "events_inserted_total",
		Help:      "Total number of click events successfully inserted into database.",
	},
)

// AnalyticsBatchErrors counts failed batch inserts.
// If this increases, check your DB health — connection pool exhaustion,
// disk full, or schema issues.
var AnalyticsBatchErrors = prometheus.NewCounter(
	prometheus.CounterOpts{
		Namespace: namespace,
		Subsystem: "analytics",
		Name:      "batch_errors_total",
		Help:      "Total number of failed batch inserts.",
	},
)

// ========================
// Cache Metrics
// ========================

// CacheRequests counts cache lookups, labeled by result: "hit" or "miss".
// cache hit ratio = rate(hit) / rate(hit + miss)
// This is arguably the most important metric for a URL shortener,
// because redirect latency depends almost entirely on whether
// the URL is in cache or requires a DB roundtrip.
var CacheRequests = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Namespace: namespace,
		Subsystem: "cache",
		Name:      "requests_total",
		Help:      "Total cache lookups, labeled by result (hit/miss).",
	},
	[]string{"result"}, // "hit" or "miss"
)

// CacheErrors counts cache infrastructure errors (Redis down, timeout, etc).
// Separate from miss — a miss is normal, an error is not.
var CacheErrors = prometheus.NewCounter(
	prometheus.CounterOpts{
		Namespace: namespace,
		Subsystem: "cache",
		Name:      "errors_total",
		Help:      "Total cache operation errors (Redis failures).",
	},
)

// ========================
// Registration
// ========================

// Register adds all metrics to the default Prometheus registry.
//
// Why a dedicated function instead of init()?
//   - Explicit is better than implicit (Go philosophy)
//   - Easier to test (you can skip registration in tests)
//   - Clear startup sequence in main.go
//
// MustRegister panics if registration fails (duplicate metric name, etc).
// This is intentional — a metric naming conflict is a programming error
// that should be caught at startup, not silently ignored at runtime.
func Register() {
	prometheus.MustRegister(
		HTTPRequestsTotal,
		HTTPRequestDuration,
		HTTPRequestsInFlight,
		AnalyticsQueueDepth,
		AnalyticsEventsDropped,
		AnalyticsBatchFlushTotal,
		AnalyticsEventsInserted,
		AnalyticsBatchErrors,
		CacheRequests,
		CacheErrors,
	)
}
