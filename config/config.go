package config

import (
	"github.com/apex/log"
	"github.com/caarlos0/env/v6"
)

// Config configuration.
type Config struct {
	RedisURL              string   `env:"REDIS_URL" envDefault:"redis://localhost:6379"`
	GitHubTokens          []string `env:"GITHUB_TOKENS"`
	GitHubPageSize        int      `env:"GITHUB_PAGE_SIZE" envDefault:"100"`
	GitHubMaxRateUsagePct int      `env:"GITHUB_MAX_RATE_LIMIT_USAGE" envDefault:"80"`
	Listen                string   `env:"LISTEN" envDefault:"127.0.0.1:3000"`
}

// Get the current Config.
func Get() (cfg Config) {
	if err := env.Parse(&cfg); err != nil { // 解析配置文件
		log.WithError(err).Fatal("failed to load config")
	}
	return
}
