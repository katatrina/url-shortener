package config

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
)

type Environment string

var (
	EnvLocal      Environment = "local"
	EnvProduction Environment = "production"
)

func (e Environment) IsValid() bool {
	switch e {
	case EnvLocal, EnvProduction:
		return true
	}

	return false
}

type ServerConfig struct {
	Port            string        `env:"SERVER_PORT" envDefault:"8080"`
	BaseURL         string        `env:"BASE_URL,required"`
	ReadTimeout     time.Duration `env:"SERVER_READ_TIMEOUT" envDefault:"10s"`
	WriteTimeout    time.Duration `env:"SERVER_WRITE_TIMEOUT" envDefault:"15s"`
	IdleTimeout     time.Duration `env:"SERVER_IDLE_TIMEOUT" envDefault:"60s"`
	ShutdownTimeout time.Duration `env:"SERVER_SHUTDOWN_TIMEOUT" envDefault:"10s"`
}

type DatabaseConfig struct {
	URL string `env:"DATABASE_URL,required"`
}

type RedisConfig struct {
	URL string `env:"REDIS_URL,required"`
}

type JWTConfig struct {
	SecretKey string        `env:"JWT_SECRET_KEY,required"`
	TTL       time.Duration `env:"JWT_TTL" envDefault:"24h"`
}

type Config struct {
	AppEnv      Environment `env:"APP_ENV,required"`
	Server      ServerConfig
	Database    DatabaseConfig
	Redis       RedisConfig
	JWT         JWTConfig
	CORSOrigins []string `env:"CORS_ORIGINS"`
	LogLevel    string   `env:"LOG_LEVEL" envDefault:"debug"`
}

func (c *Config) IsProduction() bool {
	return c.AppEnv == EnvProduction
}

// Validate .
func (c *Config) Validate() error {
	if !c.AppEnv.IsValid() {
		return fmt.Errorf("invalid APP_ENV %q (must be 'local' or 'production')", c.AppEnv)
	}

	const minJWTSecretKey = 32
	if len(c.JWT.SecretKey) < minJWTSecretKey {
		return fmt.Errorf("JWT_SECRET_KEY must be at least %d bytes, got %d", minJWTSecretKey, len(c.JWT.SecretKey))
	}

	if c.JWT.TTL <= 0 {
		return fmt.Errorf("JWT_TTL must be positive, got %v", c.JWT.TTL)
	}

	c.LogLevel = strings.ToLower(c.LogLevel)
	switch c.LogLevel {
	case "debug", "info", "warn", "error":
	default:
		return fmt.Errorf("invalid LOG_LEVEL %q", c.LogLevel)
	}

	return nil
}

// Load .
func Load() (*Config, error) {
	if err := godotenv.Load(); err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, err
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

func (c *Config) LogEffective() {
	mask := func(s string) string {
		if s == "" {
			return "<empty>"
		}

		return fmt.Sprintf("<set, len=%d>", len(s))
	}

	slog.Info("effective config",
		"APP_ENV", c.AppEnv,
		"SERVER_PORT", c.Server.Port,
		"BASE_URL", c.Server.BaseURL,
		"SERVER_READ_TIMEOUT", c.Server.ReadTimeout,
		"SERVER_WRITE_TIMEOUT", c.Server.WriteTimeout,
		"SERVER_IDLE_TIMEOUT", c.Server.IdleTimeout,
		"SERVER_SHUTDOWN_TIMEOUT", c.Server.ShutdownTimeout,
		"DATABASE_URL", mask(c.Database.URL),
		"REDIS_URL", mask(c.Redis.URL),
		"JWT_SECRET_KEY", mask(c.JWT.SecretKey),
		"JWT_TTL", c.JWT.TTL,
		"CORS_ORIGINS", c.CORSOrigins,
		"LOG_LEVEL", c.LogLevel,
	)
}
