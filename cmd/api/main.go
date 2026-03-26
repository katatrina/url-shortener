package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
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
	"github.com/katatrina/url-shortener/internal/middleware"
	"github.com/katatrina/url-shortener/internal/repository"
	"github.com/katatrina/url-shortener/internal/response"
	"github.com/katatrina/url-shortener/internal/service"
	"github.com/katatrina/url-shortener/internal/token"
	"github.com/redis/go-redis/v9"
)

func main() {
	cfg, err := config.LoadConfig(".env")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Create a context that is canceled when the process receives SIGINT or SIGTERM.
	//
	// signal.NotifyContext is cleaner than manually creating channels and calling
	// signal.Notify. The returned ctx is canceled on the first signal, and calling
	// stop() de-registers the signal handler (restoring default behavior so a
	// second Ctrl+C force-kills the process).
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// ---- Database ----
	dbCtx, dbCancel := context.WithTimeout(ctx, 10*time.Second)
	defer dbCancel()

	db, err := pgxpool.New(dbCtx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to create database pool: %v", err)
	}
	defer db.Close()

	if err = db.Ping(dbCtx); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}
	log.Println("Connected to database successfully")

	// ---- Redis ----
	redisOpts, err := redis.ParseURL(cfg.RedisURL)
	if err != nil {
		log.Fatalf("Failed to parse Redis URL: %v", err)
	}
	rdb := redis.NewClient(redisOpts)
	defer rdb.Close()

	if err = rdb.Ping(ctx).Err(); err != nil {
		log.Fatalf("Failed to ping Redis: %v", err)
	}
	log.Println("Connected to Redis successfully")

	// ---- Dependencies ----
	tokenMaker, err := token.NewJWTMaker([]byte(cfg.JWTSecret), cfg.JWTExpiry)
	if err != nil {
		log.Fatalf("Failed to create token maker: %v", err)
	}

	rateLimiter := redis_rate.NewLimiter(rdb)

	urlCache := cache.NewURLCache(rdb)
	urlRepo := repository.NewURLRepository(db)
	userRepo := repository.NewUserRepository(db)
	clickEventRepo := repository.NewClickEventRepository(db)

	svc := service.New(urlRepo, userRepo, urlCache, tokenMaker)

	// ---- Analytics Collector ----
	collector := analytics.NewCollector(clickEventRepo, analytics.DefaultCollectorConfig())
	collector.Start()

	h := handler.New(svc, collector, cfg.BaseURL)

	// ---- Router ----
	router := gin.Default()
	router.NoRoute(func(c *gin.Context) {
		response.NotFound(c, response.CodeRouteNotFound,
			"The requested endpoint does not exist")
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

		protected := v1.Group("/me/urls")
		protected.Use(middleware.Auth(tokenMaker))
		{
			protected.GET("", h.ListUserURLs)
			protected.GET("/:code", h.GetUserURL)
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
		log.Printf("Server starting on %s", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	// ---- Graceful Shutdown ----
	//
	// Block here until we receive SIGINT (Ctrl+C) or SIGTERM (Docker/K8s stop).
	// When a signal arrives, ctx is canceled, and <-ctx.Done() unblocks.
	<-ctx.Done()
	log.Println("Shutdown signal received")

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
		log.Printf("[ERROR] HTTP server shutdown error: %v", err)
	}
	log.Println("HTTP server stopped")

	// Step 2: Stop the analytics collector.
	// This must happen AFTER the HTTP server stops, because:
	// - HTTP server might still be handling in-flight requests
	// - Those requests might push events to the collector
	// - If we stop the collector first, those events would be sent to a closed channel → panic
	//
	// Sequence matters:
	// 1. Stop HTTP server (no more requests → no more events produced)
	// 2. Stop collector (drain remaining events → flush final batch)
	collector.Stop()

	// Step 3: Close infrastructure connections.
	// DB and Redis are closed by their deferred Close() calls.
	log.Println("Shutdown complete")
}
