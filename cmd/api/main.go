package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis_rate/v10"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/katatrina/url-shortener/internal/analytics"
	"github.com/katatrina/url-shortener/internal/cache"
	"github.com/katatrina/url-shortener/internal/config"
	"github.com/katatrina/url-shortener/internal/handler"
	"github.com/katatrina/url-shortener/internal/logger"
	"github.com/katatrina/url-shortener/internal/metrics"
	"github.com/katatrina/url-shortener/internal/middleware"
	"github.com/katatrina/url-shortener/internal/repository"
	"github.com/katatrina/url-shortener/internal/response"
	"github.com/katatrina/url-shortener/internal/service"
	"github.com/katatrina/url-shortener/internal/token"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/redis/go-redis/v9"
)

func main() {
	logger.Setup()

	cfg, err := config.LoadConfig(".env")
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	// ---- Metrics ----
	// Register all Prometheus metrics before starting the server.
	// Must happen before any metric is used, otherwise Inc()/Observe() panics.
	metrics.Register()

	// Create a context that is canceled when the process receives SIGINT or SIGTERM.
	//
	// signal.NotifyContext is cleaner than manually creating channels and calling
	// signal.Notify. The returned ctx is canceled on the first signal, and calling
	// stop() de-registers the signal handler (restoring default behavior so a
	// second Ctrl+C force-kills the process).
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// ---- Database ----
	dbCtx, dbCancel := context.WithTimeout(ctx, 10*time.Second)
	defer dbCancel()

	db, err := pgxpool.New(dbCtx, cfg.DatabaseURL)
	if err != nil {
		slog.Error("failed to create database pool", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	if err = db.Ping(dbCtx); err != nil {
		slog.Error("failed to ping database", "error", err)
		os.Exit(1)
	}
	slog.Info("connected to database")

	// Register DB pool metrics AFTER pool is created.
	metrics.RegisterDBPool(db)

	// ---- Redis ----
	redisOpts, err := redis.ParseURL(cfg.RedisURL)
	if err != nil {
		slog.Error("failed to parse Redis URL", "error", err)
		os.Exit(1)
	}
	rdb := redis.NewClient(redisOpts)
	defer rdb.Close()

	if err = rdb.Ping(ctx).Err(); err != nil {
		slog.Error("failed to ping Redis", "error", err)
		os.Exit(1)
	}
	slog.Info("connected to Redis")

	// ---- Dependencies ----
	tokenMaker, err := token.NewJWTMaker([]byte(cfg.JWTSecret), cfg.JWTExpiry)
	if err != nil {
		slog.Error("failed to create token maker", "error", err)
		os.Exit(1)
	}

	rateLimiter := redis_rate.NewLimiter(rdb)

	urlCache := cache.NewURLCache(rdb)
	urlRepo := repository.NewURLRepository(db)
	userRepo := repository.NewUserRepository(db)
	clickEventRepo := repository.NewClickEventRepository(db)
	statsRepo := repository.NewURLStatsRepository(db)

	// ---- Analytics ClickCollector ----
	collector := analytics.NewClickCollector(clickEventRepo, analytics.DefaultCollectorConfig())
	collector.Start()

	aggregator := analytics.NewAggregator(statsRepo, 1*time.Minute) // 1 min for dev, 5-15 min for prod
	aggregator.Start()

	svc := service.New(urlRepo, userRepo, urlCache, clickEventRepo, statsRepo, tokenMaker, collector)
	h := handler.New(svc, cfg.BaseURL)

	// ---- Router ----
	router := gin.Default()

	// Metrics middleware must be FIRST so it captures ALL requests,
	// including those rejected by auth or rate limiting.
	router.Use(middleware.Metrics())

	router.NoRoute(func(c *gin.Context) {
		response.NotFound(c, response.CodeRouteNotFound,
			"The requested endpoint does not exist")
	})

	// Prometheus metrics endpoint.
	//
	// Why gin.WrapH instead of writing our own handler?
	// promhttp.Handler() returns a standard net/http handler that formats
	// all registered metrics in Prometheus's exposition format.
	// gin.WrapH adapts it to Gin's handler signature.
	//
	// This endpoint is intentionally NOT under /api/v1 because:
	//   1. It's for infrastructure (Prometheus), not for API consumers
	//   2. By convention, /metrics is the standard path Prometheus expects
	//   3. It should NOT go through auth or rate limiting middleware
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	router.GET("/:code", h.Redirect)

	v1 := router.Group("/api/v1")
	{
		v1.POST("/shorten",
			middleware.RateLimit(rateLimiter, redis_rate.PerMinute(10)), // 10 requests per minute per IP.
			middleware.OptionalAuth(tokenMaker),
			h.ShortenURL,
		)

		v1.POST("/auth/register", h.Register)
		v1.POST("/auth/login", h.Login)

		protected := v1.Group("/me/urls")
		protected.Use(middleware.Auth(tokenMaker))
		{
			protected.GET("", h.ListUserURLs)
			protected.GET("/:code", h.GetUserURL)
			protected.GET("/:code/stats", h.GetURLStats)
			protected.DELETE("/:code", h.DeleteUserURL)
		}
	}

	// ---- Start Server ----
	//
	// We use http.Server directly instead of router.Run() because router.Run()
	// calls http.ListenAndServe() internally, which blocks forever and provides
	// no way to shut down gracefully.
	//
	// With http.Server, we can call server.Shutdown() which:
	// 1. Stops accepting new connections
	// 2. Waits for in-flight requests to complete (up to the timeout)
	// 3. Returns, allowing us to clean up other resources (collector, DB, Redis)
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", cfg.ServerPort),
		Handler: router,
	}

	// Start the server in a goroutine so it doesn't block the main goroutine.
	// The main goroutine needs to stay free to listen for shutdown signals.
	go func() {
		slog.Info("server starting", "addr", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("server failed", "error", err)
			os.Exit(1)
		}
	}()

	// ---- Graceful Shutdown ----
	//
	// Block here until we receive SIGINT (Ctrl+C) or SIGTERM (Docker/K8s stop).
	// When a signal arrives, ctx is canceled, and <-ctx.Done() unblocks.
	<-ctx.Done()
	slog.Info("shutdown signal received")

	// Restore default signal behavior. If the operator sends a second signal
	// during shutdown (double Ctrl+C), the process will exit immediately
	// instead of waiting for graceful shutdown to complete.
	stop()

	// Step 1: Stop accepting new HTTP requests and wait for in-flight ones.
	// The 10-second timeout is a safety net — if some request is stuck
	// (e.g., slow DB query), we don't wait forever.
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("HTTP server shutdown error", "error", err)
	}
	slog.Info("HTTP server stopped")

	// Step 2: Stop the analytics collector first.
	// This must happen AFTER the HTTP server stops, because:
	// - HTTP server might still be handling in-flight requests
	// - Those requests might push events to the collector
	// - If we stop the collector first, those events would be sent to a closed channel → panic
	//
	// Sequence matters:
	// 1. Stop HTTP server (no more requests → no more events produced)
	// 2. Stop collector (drain remaining events → flush final batch)
	collector.Stop()

	// Stop aggregator last — run final aggregation to capture
	// any events the collector just flushed.
	aggregator.Stop()

	// Step 3: Close infrastructure connections.
	// DB and Redis are closed by their deferred Close() calls.
	slog.Info("shutdown complete")
}
