package config

import (
	"os"

	env "github.com/joho/godotenv"
)

// Config holds application configuration settings
type Config struct {
	RunAddr    string `env:"RUN_ADDRESS" envDefault:"localhost:8080"`
	ServerAddr string `env:"SERVER_ADDRESS" envDefault:"localhost:50051"`
}

// New creates and loads a new configuration instance
func New() (Config, error) {
	cfg := Config{}

	err := env.Load()
	if err != nil && !os.IsNotExist(err) {
		return cfg, err
	}

	if cfg.RunAddr == "" {
		cfg.RunAddr = os.Getenv("RUN_ADDRESS")
	}

	if cfg.ServerAddr == "" {
		cfg.ServerAddr = os.Getenv("SERVER_ADDRESS")
	}

	return cfg, nil
}
