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

	"github.com/gin-contrib/cors"
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
	// Đọc APP_ENV TRƯỚC khi load config
	// Vì nếu config loading fail, ta cần logger đã sẵn sàng để log lỗi
	env := os.Getenv("APP_ENV")
	if env == "" {
		env = "development"
	}
	logger.Setup(env)

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
	defer rdb.Close() //nolint:errcheck

	if err = rdb.Ping(ctx).Err(); err != nil {
		slog.Error("failed to ping Redis", "error", err)
		os.Exit(1)
	}
	slog.Info("connected to Redis")

	// ---- Dependencies ----
	tokenMaker := token.NewJWTMaker(cfg.JWTSecret, cfg.JWTTTL)

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
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()

	// Order of importance for middleware — read from top to bottom:
	// 1. Metrics: measures all requests (even those rejected by subsequent middleware)
	// 2. RequestID: creates an ID, attaches it to the context — all later logs will have the ID
	// 3. Logging: logs the start/end request with the request_id
	// 4. Recovery: catches panic, prevents server crashes
	// 5. CORS: handles cross-origin
	router.Use(middleware.Metrics())
	router.Use(middleware.RequestID())
	router.Use(middleware.Logging())
	router.Use(gin.Recovery())

	// CORS must be before any route handlers.
	// AllowOrigins is loaded from CORS_ORIGINS env var (comma-separated).
	// Example: CORS_ORIGINS=https://myapp.com,https://staging.myapp.com
	router.Use(cors.New(cors.Config{
		AllowOrigins:     cfg.CORSOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"RateLimit-Limit", "RateLimit-Remaining", "RateLimit-Reset", "Retry-After"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

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

	// Liveness: "process có đang chạy không?"
	// K8s/Docker dùng cái này để quyết định có restart container không.
	// Chỉ cần return 200 — nếu process treo thì tự khắc không respond.
	router.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Readiness: "app có sẵn sàng nhận traffic không?"
	// K8s/load balancer dùng cái này để quyết định có route traffic không.
	// Check tất cả dependencies: DB, Redis.
	router.GET("/readyz", func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
		defer cancel()

		if err := db.Ping(ctx); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status": "not ready",
				"reason": "database unavailable",
			})
			return
		}

		if err := rdb.Ping(ctx).Err(); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status": "not ready",
				"reason": "redis unavailable",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{"status": "ready"})
	})

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

		protected := v1.Group("/me")
		protected.Use(middleware.Auth(tokenMaker))
		{
			urlGroup := protected.Group("/urls")
			{
				urlGroup.GET("", h.ListUserURLs)
				urlGroup.GET("/:code", h.GetUserURL)
				urlGroup.GET("/:code/stats", h.GetURLStats)
				urlGroup.DELETE("/:code", h.DeleteUserURL)
			}

			protected.GET("/profile", h.GetUserProfile)
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
