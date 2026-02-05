package config

import (
	"github.com/caarlos0/env/v6"
	"github.com/charmbracelet/log"
)

// Config configuration.
type Config struct {
	RedisURL              string   `env:"REDIS_URL" envDefault:"redis://localhost:6379"`
	GitHubTokens          []string `env:"GITHUB_TOKENS"`
	GitHubPageSize        int      `env:"GITHUB_PAGE_SIZE" envDefault:"100"`
	GitHubMaxRateUsagePct int      `env:"GITHUB_MAX_RATE_LIMIT_USAGE" envDefault:"80"`
	GitHubMaxSamplePages  int      `env:"GITHUB_MAX_SAMPLE_PAGES" envDefault:"15"`
	Listen                string   `env:"LISTEN" envDefault:"127.0.0.1:3000"`
}

// Get the current Config.
func Get() (cfg Config) {
	if err := env.Parse(&cfg); err != nil {
		log.Fatal("failed to load config", "err", err)
	}
	return
}
