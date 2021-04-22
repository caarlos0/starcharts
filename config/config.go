package config

import (
	"github.com/apex/log"
	"github.com/caarlos0/env/v6"
)

// Config configuration.
type Config struct {
	RedisURL    string `env:"REDIS_URL" envDefault:"redis://localhost:6379"`
	GitHubToken string `env:"GITHUB_TOKEN"`
	Port        string `env:"PORT" envDefault:"3000"`
}

// Get the current Config.
func Get() (cfg Config) {
	if err := env.Parse(&cfg); err != nil {
		log.WithError(err).Fatal("failed to load config")
	}
	return
}
