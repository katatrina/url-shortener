package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
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

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	db, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to create database pool: %v", err)
	}
	defer db.Close()

	if err = db.Ping(ctx); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}
	log.Println("Connected to database successfully")

	// Redis
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

	tokenMaker, err := token.NewJWTMaker([]byte(cfg.JWTSecret), cfg.JWTExpiry)
	if err != nil {
		log.Fatalf("Failed to create token maker: %v", err)
	}

	urlCache := cache.NewURLCache(rdb)

	urlRepo := repository.NewURLRepository(db)
	userRepo := repository.NewUserRepository(db)
	svc := service.New(urlRepo, userRepo, urlCache, tokenMaker)
	h := handler.New(svc, cfg.BaseURL)

	router := gin.Default()
	router.NoRoute(func(c *gin.Context) {
		response.NotFound(c, response.CodeRouteNotFound,
			"The requested endpoint does not exist")
	})

	router.GET("/:code", h.Redirect)

	v1 := router.Group("/api/v1")
	{
		// Public: shorten URL (works with or without auth)
		v1.POST("/shorten", middleware.OptionalAuth(tokenMaker), h.ShortenURL)

		// Auth
		v1.POST("/auth/register", h.Register)
		v1.POST("/auth/login", h.Login)

		// Protected: manage auth user URLs
		protected := v1.Group("/me/urls")
		protected.Use(middleware.Auth(tokenMaker))
		{
			protected.GET("", h.ListUserURLs)
			protected.GET("/:code", h.GetUserURL)
			protected.DELETE("/:code", h.DeleteUserURL)
		}
	}

	addr := fmt.Sprintf(":%s", cfg.ServerPort)
	log.Printf("Server starting on %s", addr)
	if err = router.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
