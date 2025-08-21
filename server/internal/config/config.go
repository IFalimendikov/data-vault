package config

import (
	"os"

	env "github.com/joho/godotenv"
)

type Config struct {
	ServerAddr    string `env:"RUN_ADDRESS" envDefault:"localhost:8080"`
	DatabaseURI   string `env:"DATABASE_URI"`
	JWTSecret     string `env:"JWT_SECRET" envDefault:"123"`
	EncryptionKey string `env:"ENCRYPTION_KEY" envDefault:"123"`
}

func New() (Config, error) {
	cfg := Config{}

	err := env.Load()
	if err != nil && !os.IsNotExist(err) {
		return cfg, err
	}

	if cfg.ServerAddr == "" {
		cfg.ServerAddr = os.Getenv("RUN_ADDRESS")
	}

	if cfg.DatabaseURI == "" {
		cfg.DatabaseURI = os.Getenv("DATABASE_URI")
	}

	if cfg.JWTSecret == "" {
		cfg.JWTSecret = os.Getenv("JWT_SECRET")
	}

	if cfg.EncryptionKey == "" {
		cfg.EncryptionKey = os.Getenv("ENCRYPTION_KEY")
	}

	return cfg, nil
}
