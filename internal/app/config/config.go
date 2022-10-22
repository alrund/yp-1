package config

import (
	"log"
	"os"

	"github.com/alrund/yp-1/internal/app/flags"
	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	ServerAddress     string `env:"SERVER_ADDRESS" env-default:"localhost:8080" json:"server_address"`
	BaseURL           string `env:"BASE_URL" env-default:"http://localhost:8080/" json:"base_url"`
	GrpcServerAddress string `env:"GRPC_SERVER_ADDRESS" env-default:"localhost:9090" json:"grpc_server_address"`
	FileStoragePath   string `env:"FILE_STORAGE_PATH" json:"file_storage_path"`
	DatabaseDsn       string `env:"DATABASE_DSN" json:"database_dsn"` // postgres://dev:dev@localhost:5432/dev
	CipherPass        string `env:"CIPHER_PASSWORD" env-default:"J53RPX6" json:"-"`
	EnableHTTPS       bool   `env:"ENABLE_HTTPS" json:"enable_https"`
	CertFile          string `env:"CERT_FILE" json:"cert_file"`
	KeyFile           string `env:"KEY_FILE" json:"key_file"`
	TrustedSubnet     string `env:"TRUSTED_SUBNET" json:"trusted_subnet"`
}

// GetConfig returns configuration data with priority order: flags, env, config.
// Each item takes precedence over the next item.
func GetConfig() *Config {
	cfg := &Config{}

	f := flags.NewFlags()

	var configFile string
	configFile, ok := os.LookupEnv("CONFIG")
	if !ok && f.C != "" {
		configFile = f.C
	}

	if configFile != "" {
		err := cleanenv.ReadConfig(configFile, cfg)
		if err != nil {
			log.Fatal(err)
		}
	}

	err := cleanenv.ReadEnv(cfg)
	if err != nil {
		log.Fatal(err)
	}

	ReadFlags(f, cfg)
	return cfg
}

func ReadFlags(f *flags.Flags, cfg *Config) {
	if f.A != flags.NotAvailable {
		cfg.ServerAddress = f.A
	}
	if f.GA != flags.NotAvailable {
		cfg.GrpcServerAddress = f.GA
	}
	if f.B != flags.NotAvailable {
		cfg.BaseURL = f.B
	}
	if f.F != flags.NotAvailable {
		cfg.FileStoragePath = f.F
	}
	if f.D != flags.NotAvailable {
		cfg.DatabaseDsn = f.D
	}
	if f.S != flags.NotAvailable {
		cfg.EnableHTTPS = f.S != ""
	}
	if f.Crt != flags.NotAvailable {
		cfg.CertFile = f.Crt
	}
	if f.Key != flags.NotAvailable {
		cfg.KeyFile = f.Key
	}
	if f.T != flags.NotAvailable {
		cfg.TrustedSubnet = f.T
	}
}
