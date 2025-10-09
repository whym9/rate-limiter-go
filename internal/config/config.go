package config

import (
	"fmt"

	"github.com/caarlos0/env/v10"
)

type Config struct {
	HTTPAddr      string `env:"HTTP_ADDRESS" envDefault:":1234"`
	RateLimit     int    `env:"RATE_LIMIT" envDefault:"100"`
	WindowSec     int    `env:"WINDOW_SEC" envDefault:"0"`
	RedisAddress  string `env:"REDIS_ADDRESS"`
	RedisPassword string `env:"REDIS_PASSWORD"`
	RedisDB       int    `env:"REDIS_DB" envDefault:"0"`
}

func GetConfig() (*Config, error) {
	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (c *Config) Validate() error {
	if c.RateLimit <= 0 {
		return fmt.Errorf("rate limit must be a positive number")
	}
	if c.WindowSec <= 0 {
		return fmt.Errorf("window sec must be a positive number")
	}

	return nil
}
