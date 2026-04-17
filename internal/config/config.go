package config

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
)

type Environment string

const (
	EnvLocal      Environment = "local"
	EnvProduction Environment = "production"
)

type Config struct {
	AppEnv      Environment   `env:"APP_ENV,required"`
	ServerPort  string        `env:"SERVER_PORT,required"`
	BaseURL     string        `env:"BASE_URL,required"`
	DatabaseURL string        `env:"DATABASE_URL,required"`
	RedisURL    string        `env:"REDIS_URL,required"`
	JWTSecret   string        `env:"JWT_SECRET,required"`
	JWTTTL      time.Duration `env:"JWT_TTL,required"`
	CORSOrigins []string      `env:"CORS_ORIGINS" envSeparator:","`
}

func (c Config) LogEffective() {
	mask := func(s string) string {
		if s == "" {
			return "<empty>"
		}
		return fmt.Sprintf("<set, len=%d>", len(s))
	}

	slog.Info("effective config",
		"APP_ENV", c.AppEnv,
		"SERVER_PORT", c.ServerPort,
		"BASE_URL", c.BaseURL,
		"DATABASE_URL", mask(c.DatabaseURL),
		"REDIS_URL", mask(c.RedisURL),
		"JWT_SECRET", mask(c.JWTSecret),
		"JWT_TTL", c.JWTTTL,
		"CORS_ORIGINS", c.CORSOrigins,
	)
}

func Load(path string) (Config, error) {
	if path != "" {
		if err := godotenv.Load(path); err != nil {
			if errors.Is(err, os.ErrNotExist) {
				slog.Info("no .env file found, using environment variables only", "path", path)
			} else {
				return Config{}, fmt.Errorf("load .env file: %w", err)
			}
		}
	}

	var cfg Config
	if err := env.Parse(&cfg); err != nil {
		return Config{}, fmt.Errorf("parse config: %w", err)
	}

	return cfg, nil
}
