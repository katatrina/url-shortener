package config

import (
	"fmt"
	"log/slog"
	"os"
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
	// no trailing path), e.g. "https://short.example".
	ShortURLBase   string   `env:"SHORT_URL_BASE,required"`
	AllowedOrigins []string `env:"ALLOWED_ORIGINS,required"`

	// RedirectHost is the bare host used to match incoming requests against the
	// redirect service. Configured separately from ShortURLBase; keep the two in
	// sync (the host here must equal the host in SHORT_URL_BASE).
	RedirectHost string `env:"REDIRECT_HOST,required"`

	// TrustedPlatform names the header the hosting platform's edge proxy sets
	// with the real client IP (e.g. "Fly-Client-IP" on Fly.io). When set, Gin
	// reads ClientIP from it instead of the client-spoofable X-Forwarded-For
	// chain. Left empty in local/dev where no trusted proxy sits in front.
	TrustedPlatform string `env:"TRUSTED_PLATFORM"`

	// MaxLinksPerUser caps how many links a single user may own. It's an
	// abuse fuse for a public, billing-free service, not a product tier.
	// Defaulted (not required) because it's a policy knob with an obvious
	// sane value — a missing value shouldn't stop the app from booting.
	MaxLinksPerUser int `env:"MAX_LINKS_PER_USER" envDefault:"10"`
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

	if c.MaxLinksPerUser < 1 {
		return fmt.Errorf("MAX_LINKS_PER_USER must be >= 1, got %d", c.MaxLinksPerUser)
	}

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

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func (c *Config) Log() {
	slog.Info("config loaded",
		"APP_ENV", c.AppEnv,
		"LOG_LEVEL", c.LogLevel,
		"SHORT_URL_BASE", c.ShortURLBase,
		"REDIRECT_HOST", c.RedirectHost,
		"TRUSTED_PLATFORM", c.TrustedPlatform,
		"ALLOWED_ORIGINS", c.AllowedOrigins,
		"JWT_TTL", c.JWTTTL,
		"MAX_LINKS_PER_USER", c.MaxLinksPerUser,
	)
}
