package executor

import (
	"log"
	"time"

	"github.com/caarlos0/env"
)

type Config struct {
	BackendHost  string        `env:"BACKEND_API_HOST,required"`
	BackendPort  string        `env:"BACKEND_API_PORT,required"`
	PollInterval time.Duration `env:"POLL_INTERVAL,required"`
}

func NewConfig() *Config {
	cfg := Config{}
	if err := env.Parse(&cfg); err != nil {
		log.Fatalf("Failed to parse env: %v", err)
	}

	return &cfg
}
