package config

import (
	"errors"
	"log/slog"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	AppEnv      string        `mapstructure:"APP_ENV"`
	ServerPort  string        `mapstructure:"SERVER_PORT"`
	BaseURL     string        `mapstructure:"BASE_URL"`
	DatabaseURL string        `mapstructure:"DATABASE_URL"`
	RedisURL    string        `mapstructure:"REDIS_URL"`
	JWTSecret   string        `mapstructure:"JWT_SECRET"`
	JWTTTL      time.Duration `mapstructure:"JWT_TTL"`
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

func LoadConfig(path string) (*Config, error) {
	viper.AutomaticEnv()

	// .env file là optional — chỉ dùng cho local dev
	// Production dùng env vars (từ Docker, K8s, systemd,...)
	if path != "" {
		viper.SetConfigFile(path)
		if err := viper.ReadInConfig(); err != nil {
			// Không có file .env thì không sao — env vars vẫn hoạt động
			slog.Info("no .env file found, using environment variables only",
				"path", path)
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
