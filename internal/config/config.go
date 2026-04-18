package config

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
)

type Environment string

const (
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

type Config struct {
	AppEnv Environment `env:"APP_ENV" envDefault:"local"`

	Server   ServerConfig
	Database DatabaseConfig
	Redis    RedisConfig
	JWT      JWTConfig
	CORS     CORSConfig
	Logger   LoggerConfig
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
	Secret string        `env:"JWT_SECRET,required"`
	TTL    time.Duration `env:"JWT_TTL" envDefault:"24h"`
}

type CORSConfig struct {
	Origins []string `env:"CORS_ORIGINS" envSeparator:","`
}

type LoggerConfig struct {
	Level string `env:"LOG_LEVEL"`
}

func (c *Config) IsProduction() bool { return c.AppEnv == EnvProduction }
func (c *Config) IsLocal() bool      { return c.AppEnv == EnvLocal }

// Validate checks constraints that the `env` tag cannot handle.
func (c *Config) Validate() error {
	if !c.AppEnv.IsValid() {
		return fmt.Errorf("invalid APP_ENV %q (must be 'local' or 'production')", c.AppEnv)
	}

	if _, err := strconv.Atoi(c.Server.Port); err != nil {
		return fmt.Errorf("invalid SERVER_PORT %q: must be a number", c.Server.Port)
	}

	// HS256 requires secret key >= 32 bytes (RFC 7518).
	const minJWTSecretLen = 32
	if len(c.JWT.Secret) < minJWTSecretLen {
		return fmt.Errorf("JWT_SECRET must be at least %d bytes, got %d", minJWTSecretLen, len(c.JWT.Secret))
	}

	if c.JWT.TTL <= 0 {
		return fmt.Errorf("JWT_TTL must be positive, got %s", c.JWT.TTL)
	}

	if c.Logger.Level != "" {
		switch strings.ToLower(c.Logger.Level) {
		case "debug", "info", "warn", "error":
		default:
			return fmt.Errorf("invalid LOG_LEVEL %q", c.Logger.Level)
		}
	}

	return nil
}

// Load đọc config từ .env (nếu có) + env vars, parse và validate.
//
// Flow:
//  1. Cố gắng load file .env ở CWD. Không có thì bỏ qua — ở production
//     (container, fly.io) env vars được inject trực tiếp, không cần file.
//  2. Parse tất cả env vars vào struct.
//  3. Validate các ràng buộc nghiệp vụ.
func Load() (*Config, error) {
	if err := godotenv.Load(); err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, fmt.Errorf("load .env: %w", err)
	}

	var cfg Config
	if err := env.Parse(&cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("validate config: %w", err)
	}

	return &cfg, nil
}

// LogEffective in ra config đã load để debug, che các giá trị nhạy cảm.
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
		"JWT_SECRET", mask(c.JWT.Secret),
		"JWT_TTL", c.JWT.TTL,
		"CORS_ORIGINS", c.CORS.Origins,
	)
}
