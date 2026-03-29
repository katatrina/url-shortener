package middleware

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/katatrina/url-shortener/internal/metrics"
)

// Metrics returns a Gin middleware that records HTTP metrics for every request.
//
// It captures three things:
//   1. Request count (Counter) — how many requests, broken down by method/path/status
//   2. Request duration (Histogram) — how long each request took
//   3. In-flight requests (Gauge) — how many requests are being handled right now
//
// This middleware should be registered BEFORE all routes so it captures everything.
// The order in Gin's middleware chain matters — earlier middleware wraps later ones.
func Metrics() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Record that a new request has started.
		metrics.HTTPRequestsInFlight.Inc()

		start := time.Now()

		// c.Next() passes control to the next handler in the chain.
		// Everything after c.Next() runs AFTER the handler has finished.
		// This is how we measure the total duration including all handlers
		// and middleware that come after us.
		c.Next()

		// Request is done — update metrics.
		duration := time.Since(start).Seconds()
		status := strconv.Itoa(c.Writer.Status())

		// Why c.FullPath() instead of c.Request.URL.Path?
		//
		// c.Request.URL.Path returns the ACTUAL path: "/aB3kX9m"
		// c.FullPath() returns the ROUTE PATTERN: "/:code"
		//
		// If we used the actual path, every unique short code would create
		// a new time series. With millions of URLs, that's millions of series
		// → Prometheus runs out of memory → everything dies.
		//
		// By using the route pattern, we get one series per route,
		// regardless of how many URLs exist. This is the #1 mistake
		// people make when instrumenting HTTP metrics.
		path := c.FullPath()
		if path == "" {
			// FullPath is empty for unmatched routes (404 from NoRoute handler).
			// Group them under a single label to avoid cardinality explosion
			// from random paths hitting the server (bots, scanners, etc).
			path = "unmatched"
		}

		method := c.Request.Method

		metrics.HTTPRequestsTotal.WithLabelValues(method, path, status).Inc()
		metrics.HTTPRequestDuration.WithLabelValues(method, path).Observe(duration)
		metrics.HTTPRequestsInFlight.Dec()
	}
}
