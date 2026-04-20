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
	"github.com/katatrina/url-shortener/internal/middleware"
	"github.com/katatrina/url-shortener/internal/repository"
	"github.com/katatrina/url-shortener/internal/response"
	"github.com/katatrina/url-shortener/internal/service"
	"github.com/katatrina/url-shortener/internal/token"
	"github.com/redis/go-redis/v9"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		// Logger chưa setup → dùng default handler. Lỗi config là fatal,
		// thà log thô mà chạy chứ đừng cố setup logger khi config hỏng.
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	logger.Setup(cfg)
	cfg.LogEffective()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// ---- Database ----
	dbCtx, dbCancel := context.WithTimeout(ctx, 10*time.Second)
	defer dbCancel()

	db, err := pgxpool.New(dbCtx, cfg.Database.URL)
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

	// ---- Redis ----
	redisOpts, err := redis.ParseURL(cfg.Redis.URL)
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
	tokenMaker := token.NewJWTMaker(cfg.JWT.SecretKey, cfg.JWT.TTL)

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
	h := handler.New(svc, cfg.Server.BaseURL)

	// ---- Router ----
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()

	router.Use(middleware.RequestID())
	router.Use(middleware.Logger())
	router.Use(gin.Recovery())

	router.Use(cors.New(cors.Config{
		AllowOrigins:     cfg.CORSOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "X-Request-ID"},
		ExposeHeaders:    []string{"RateLimit-Limit", "RateLimit-Remaining", "RateLimit-Reset", "Retry-After", "X-Request-ID"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	router.NoRoute(func(c *gin.Context) {
		response.NotFound(c, response.CodeRouteNotFound,
			"The requested endpoint does not exist")
	})

	router.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

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

	// ---- Start HTTP Server ----

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.Server.Port),
		Handler:      router,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	go func() {
		slog.Info("server starting", "addr", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("server failed", "error", err)
			os.Exit(1)
		}
	}()

	// ---- Graceful Shutdown ----

	<-ctx.Done()
	slog.Info("shutdown signal received")

	stop()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("HTTP server shutdown error", "error", err)
	}
	slog.Info("HTTP server stopped")

	collector.Stop()

	aggregator.Stop()

	slog.Info("shutdown complete")
}
