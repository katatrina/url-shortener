package config

import (
	"errors"
	"log/slog"
	"os"
	"time"

	"github.com/spf13/viper"
)

type Environment string

const (
	EnvLocal      Environment = "local"
	EnvProduction Environment = "production"
)

type Config struct {
	AppEnv      Environment   `mapstructure:"APP_ENV"`
	ServerPort  string        `mapstructure:"SERVER_PORT"`
	BaseURL     string        `mapstructure:"BASE_URL"`
	DatabaseURL string        `mapstructure:"DATABASE_URL"`
	RedisURL    string        `mapstructure:"REDIS_URL"`
	JWTSecret   string        `mapstructure:"JWT_SECRET"`
	JWTTTL      time.Duration `mapstructure:"JWT_TTL"`
	CORSOrigins []string      `mapstructure:"CORS_ORIGINS"`
}

func (c Config) Validate() error {
	if c.AppEnv == "" {
		return errors.New("APP_ENV is required")
	}
	if c.ServerPort == "" {
		return errors.New("SERVER_PORT is required")
	}
	if c.BaseURL == "" {
		return errors.New("BASE_URL is required")
	}
	if c.DatabaseURL == "" {
		return errors.New("DATABASE_URL is required")
	}
	if c.JWTSecret == "" {
		return errors.New("JWT_SECRET is required")
	}
	if c.RedisURL == "" {
		return errors.New("REDIS_URL is required")
	}

	return nil
}

// LogEffective logs the effective configuration at startup, masking secrets.
func (c Config) LogEffective() {
	mask := func(s string) string {
		if len(s) <= 8 {
			return "***"
		}
		return s[:8] + "***"
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

func LoadConfig(path string) (*Config, error) {
	viper.AutomaticEnv()

	// BindEnv "đăng ký" từng key với Viper để nó biết cần đọc env var nào.
	//
	// Tại sao cần BindEnv khi đã có AutomaticEnv?
	//   AutomaticEnv() chỉ hoạt động với viper.Get("KEY") — gọi trực tiếp.
	//   Unmarshal() chỉ unmarshal những key mà Viper đã "biết" (từ file config,
	//   SetDefault, hoặc BindEnv). Nếu không có file .env (production), Viper
	//   không biết key nào tồn tại → Unmarshal ra struct rỗng.
	//
	// Local dev (có .env): Viper đọc file → biết tất cả keys → Unmarshal OK.
	// Docker/K8s (không có .env): Không có BindEnv → Viper không biết keys → fail.
	viper.BindEnv("APP_ENV")
	viper.BindEnv("SERVER_PORT")
	viper.BindEnv("BASE_URL")
	viper.BindEnv("DATABASE_URL")
	viper.BindEnv("REDIS_URL")
	viper.BindEnv("JWT_SECRET")
	viper.BindEnv("JWT_TTL")
	viper.BindEnv("CORS_ORIGINS")

	// .env file là optional — chỉ dùng cho local dev
	// Production dùng env vars (từ Docker, K8s, systemd,...)
	if path != "" {
		viper.SetConfigFile(path)
		if err := viper.ReadInConfig(); err != nil {
			if errors.Is(err, os.ErrNotExist) {
				// Không có file .env thì không sao — env vars vẫn hoạt động
				slog.Info("no .env file found, using environment variables only", "path", path)
			} else {
				// File .env tồn tại nhưng sai format (syntax error) -> văng lỗi để dev biết
				return nil, err
			}
		}
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return &cfg, nil
}
