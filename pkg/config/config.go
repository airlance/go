package config

import (
	"context"
	"fmt"
	"strings"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
	"github.com/sirupsen/logrus"
)

type Config struct {
	DB      DatabaseConfig `envconfig:"DB"`
	Logging LoggingConfig  `envconfig:"LOG"`
	Auth    AuthConfig     `envconfig:"AUTH"`
	Server  ServerConfig   `envconfig:"SERVER"`
}

type DatabaseConfig struct {
	DSN string `envconfig:"DSN" required:"true"`
}

type LoggingConfig struct {
	Level string `envconfig:"LEVEL" default:"debug"`
}

type AuthConfig struct {
	APITokens string `envconfig:"API_TOKENS" required:"false"`
}

type ServerConfig struct {
	Port string `envconfig:"PORT" default:"8080"`
}

func (a *AuthConfig) GetTokens() []string {
	if a.APITokens == "" {
		return []string{}
	}

	tokens := strings.Split(a.APITokens, ",")
	result := make([]string, 0, len(tokens))

	for _, token := range tokens {
		token = strings.TrimSpace(token)
		if token != "" {
			result = append(result, token)
		}
	}

	return result
}

func Init(ctx context.Context) *Config {
	cfg, err := loadConfig(ctx)
	if err != nil {
		panic(err)
	}

	return cfg
}

func loadConfig(ctx context.Context) (*Config, error) {
	if ctx == nil {
		panic("context must not be nil")
	}

	if err := godotenv.Load(); err != nil {
		logrus.Warn("No .env file found, using environment variables")
	}

	var cfg Config

	if err := envconfig.Process("", &cfg); err != nil {
		return nil, fmt.Errorf("failed to process config: %w", err)
	}

	return &cfg, nil
}
