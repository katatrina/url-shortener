package config

import (
	"fmt"
	"log/slog"
	"net/url"
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

func (e Environment) IsProduction() bool {
	return e == EnvProduction
}

type Config struct {
	AppEnv      Environment   `env:"APP_ENV,required"`
	DatabaseURL string        `env:"DATABASE_URL,required"`
	LogLevel    string        `env:"LOG_LEVEL,required"`
	JWTSecret   string        `env:"JWT_SECRET,required"`
	JWTTTL      time.Duration `env:"JWT_TTL,required"`

	// ShortURLBase is the public origin short links are built on (scheme + host,
	// no trailing path), e.g. "https://short.example". It is the single source of
	// truth: RedirectHost below is derived from it, never configured separately.
	ShortURLBase   string   `env:"SHORT_URL_BASE,required"`
	AllowedOrigins []string `env:"ALLOWED_ORIGINS,required"`

	// RedirectHost is the bare, lowercased host used to match incoming requests
	// against the redirect service. Derived from ShortURLBase in Load.
	RedirectHost string `env:"-"`
}

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

// resolveShortURLBase validates SHORT_URL_BASE as a bare origin (scheme + host,
// no path/query/fragment) and derives RedirectHost for request routing. Both
// consumers thus read from one normalized source instead of transforming the
// raw value independently.
func (c *Config) resolveShortURLBase() error {
	c.ShortURLBase = strings.TrimRight(c.ShortURLBase, "/")

	u, err := url.Parse(c.ShortURLBase)
	if err != nil {
		return fmt.Errorf("invalid SHORT_URL_BASE %q: %w", c.ShortURLBase, err)
	}

	if u.Scheme != "http" && u.Scheme != "https" {
		return fmt.Errorf("SHORT_URL_BASE %q must start with http:// or https://", c.ShortURLBase)
	}
	if u.Host == "" {
		return fmt.Errorf("SHORT_URL_BASE %q is missing a host", c.ShortURLBase)
	}
	if u.Path != "" || u.RawQuery != "" || u.Fragment != "" {
		return fmt.Errorf("SHORT_URL_BASE %q must be a bare origin (no path, query, or fragment)", c.ShortURLBase)
	}

	c.RedirectHost = strings.ToLower(u.Hostname())

	return nil
}

func Load() (*Config, error) {
	if _, err := os.Stat(".env"); err == nil {
		if err := godotenv.Load(); err != nil {
			return nil, err
		}
	}

	var cfg Config
	if err := env.Parse(&cfg); err != nil {
		return nil, err
	}

	if err := cfg.resolveShortURLBase(); err != nil {
		return nil, err
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func (c *Config) Log() {
	slog.Info("current config",
		"APP_ENV", c.AppEnv,
		"LOG_LEVEL", c.LogLevel,
	)
}
