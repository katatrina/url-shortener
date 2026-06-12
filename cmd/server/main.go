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

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/katatrina/url-shortener/internal/config"
	"github.com/katatrina/url-shortener/internal/httpserver"
	"github.com/katatrina/url-shortener/internal/logger"
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

	logger.Setup(cfg)
	cfg.Log()

	db, err := pgxpool.New(context.Background(), cfg.DatabaseURL)
	if err != nil {
		return fmt.Errorf("create db pool: %w", err)
	}
	defer db.Close()

	if err := db.Ping(context.Background()); err != nil {
		return fmt.Errorf("ping db: %w", err)
	}
	slog.Info("connected to db")

	r := httpserver.NewRouter(db)

	srv := &http.Server{
		Addr:    ":8080",
		Handler: r,

		// Thời gian tối đa server chờ đọc xong header từ client,
		// chống client gửi header nhỏ giọt (Slowloris kinh điển).
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
		slog.Info("shutting down gracefully...")
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		_ = srv.Close()
		slog.Warn("graceful shutdown failed", "error", err)
	} else {
		slog.Info("server stopped")
	}

	return nil
}
