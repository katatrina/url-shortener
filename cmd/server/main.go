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

	_ "time/tzdata"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/katatrina/url-shortener/internal/click"
	"github.com/katatrina/url-shortener/internal/config"
	"github.com/katatrina/url-shortener/internal/link"
	"github.com/katatrina/url-shortener/internal/logger"
	"github.com/katatrina/url-shortener/internal/router"
	"github.com/katatrina/url-shortener/internal/token"
	"github.com/katatrina/url-shortener/internal/user"
)

const (
	clickBufferSize    = 1024
	clickBatchSize     = 100
	clickFlushInterval = 5 * time.Second

	geoipDBPath = "geoip/dbip-country-lite.mmdb"
)

func main() {
	if err := run(); err != nil {
		slog.Error("application error", "error", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	logger.Setup(cfg.LogLevel, cfg.AppEnv.IsProduction())
	cfg.Log()

	db, err := pgxpool.New(context.Background(), cfg.DatabaseURL)
	if err != nil {
		return fmt.Errorf("create db pool: %w", err)
	}
	defer func() {
		slog.Info("closing db pool...")
		db.Close()
		slog.Info("db pool closed")
	}()

	if err := db.Ping(context.Background()); err != nil {
		return fmt.Errorf("ping db: %w", err)
	}
	slog.Info("connected to db")

	countryResolver := click.CountryResolver(click.NoopResolver{})
	if geo, err := click.NewMMDBResolver(geoipDBPath); err != nil {
		slog.Warn("geoip disabled: cannot open database",
			slog.String("path", geoipDBPath),
			slog.Any("error", err),
		)
	} else {
		countryResolver = geo
		defer func() {
			if err := geo.Close(); err != nil {
				slog.Warn("closing geoip database failed", slog.Any("error", err))
			}
		}()
		slog.Info("geoip enabled", "path", geoipDBPath)
	}

	clickPipeline := click.NewPipeline(click.NewWriter(db), countryResolver,
		clickBufferSize, clickBatchSize, clickFlushInterval)

	pipelineCtx, stopPipeline := context.WithCancel(context.Background())
	defer stopPipeline()
	pipelineDone := make(chan struct{})
	go func() {
		clickPipeline.Run(pipelineCtx)
		close(pipelineDone)
	}()

	tokenIssuer := token.NewIssuer(cfg.JWTSecret, cfg.JWTTTL)
	userHandler := user.NewHandler(user.NewService(user.NewRepository(db), tokenIssuer))
	linkSvc := link.NewService(link.NewRepository(db), click.NewStatsRepository(db), cfg.MaxLinksPerUser)
	linkHandler := link.NewHandler(linkSvc, cfg.ShortURLBase, clickPipeline)

	r := router.New(cfg, userHandler, linkHandler, tokenIssuer)

	srv := &http.Server{
		Addr:    ":8080",
		Handler: r,

		ReadHeaderTimeout: 5 * time.Second,
	}

	serverErr := make(chan error)
	go func() {
		slog.Info("server starting", "addr", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
		}
	}()

	notifyCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	select {
	case err := <-serverErr:
		return fmt.Errorf("start server: %w", err)
	case <-notifyCtx.Done():
		stop()
		slog.Info("shutting down server gracefully...")
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Warn("graceful shutdown failed", "error", err)
	} else {
		slog.Info("server stopped")
	}

	slog.Info("draining click pipeline...")
	stopPipeline()
	<-pipelineDone
	slog.Info("click pipeline drained")

	return nil
}
