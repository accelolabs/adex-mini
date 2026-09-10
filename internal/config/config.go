package config

import (
	"fmt"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	HTTPAddr        string        `env:"HTTP_ADDR" env-default:":8080"`
	DatabaseURL     string        `env:"DATABASE_URL" env-required:"true"`
	AuctionTimeout  time.Duration `env:"AUCTION_TIMEOUT" env-default:"200ms"`
	ShutdownTimeout time.Duration `env:"SHUTDOWN_TIMEOUT" env-default:"5s"`
	LogLevel        string        `env:"LOG_LEVEL" env-default:"info"`
}

func MustLoad() *Config {
	var cfg Config
	if err := cleanenv.ReadEnv(&cfg); err != nil {
		fmt.Fprintf(os.Stderr, "load config: %v\n", err)
		os.Exit(1)
	}

	return &cfg
}
