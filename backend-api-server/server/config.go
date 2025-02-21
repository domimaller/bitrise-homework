package server

import (
	"log"

	"github.com/caarlos0/env"
)

type Config struct {
	Port       string `env:"SERVER_PORT,required"`
	DBHost     string `env:"DB_HOST"`
	DBPort     string `env:"DB_PORT"`
	DBUser     string `env:"DB_USER"`
	DBPassword string `env:"DB_PASSWORD"`
	DBName     string `env:"DB_NAME"`
}

func NewConfig() *Config {
	cfg := Config{}
	if err := env.Parse(&cfg); err != nil {
		log.Fatalf("Failed to parse env: %v", err)
	}

	return &cfg
}
