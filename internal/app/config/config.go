package config

import (
	"log"

	"github.com/caarlos0/env/v6"
)

type Config struct {
	ServerAddress   string `env:"SERVER_ADDRESS" envDefault:"localhost:8080"`
	BaseURL         string `env:"BASE_URL" envDefault:"http://localhost:8080/"`
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
	DatabaseDsn     string `env:"DATABASE_DSN"` // postgres://dev:dev@localhost:5432/dev
	CipherPass      string `env:"CIPHER_PASSWORD" envDefault:"J53RPX6"`
}

// GetConfig returns configuration data.
func GetConfig() *Config {
	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		log.Fatal(err)
	}
	return cfg
}
