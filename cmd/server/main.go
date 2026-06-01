package main

import (
	"log/slog"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/katatrina/url-shortener/internal/config"
	"github.com/katatrina/url-shortener/internal/logger"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	logger.Setup(cfg)
	cfg.Log()

	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	r.GET("/healthz", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	if err := r.Run(); err != nil {
		slog.Error("failed to run server", "error", err)
	}
}
