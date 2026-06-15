package config

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
)

type Environment string

var (
	EnvLocal      Environment = "local"
	EnvProduction Environment = "production"
)

// IsValid .
func (e Environment) IsValid() bool {
	switch e {
	case EnvLocal, EnvProduction:
		return true
	}

	return false
}

func (e Environment) IsProduction() bool {
	return e == EnvProduction
}

// Config .
type Config struct {
	AppEnv      Environment `env:"APP_ENV,required"`
	DatabaseURL string      `env:"DATABASE_URL,required"`
	LogLevel    string      `env:"LOG_LEVEL,required"`
}

// Validate .
func (c *Config) Validate() error {
	if !c.AppEnv.IsValid() {
		return fmt.Errorf("invalid APP_ENV %q (must be 'local' or 'production')", c.AppEnv)
	}

	switch c.LogLevel {
	case "debug", "info", "warn", "error":
	default:
		return fmt.Errorf("invalid LOG_LEVEL %q", c.LogLevel)
	}

	return nil
}

// Load .
func Load() (*Config, error) {
	// Only load .env for local development; in production config comes from
	// real environment variables (e.g. -e flags, compose, k8s secrets).
	if _, err := os.Stat(".env"); err == nil {
		if err := godotenv.Load(); err != nil {
			return nil, err
		}
	}

	var cfg Config
	if err := env.Parse(&cfg); err != nil {
		return nil, err
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// Log .
func (c *Config) Log() {
	slog.Info("current config",
		"APP_ENV", c.AppEnv,
		"LOG_LEVEL", c.LogLevel,
	)
}
