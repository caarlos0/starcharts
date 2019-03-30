package github

import (
	"github.com/caarlos0/starcharts/config"
	"github.com/caarlos0/starcharts/internal/cache"
	"github.com/prometheus/client_golang/prometheus"
)

// GitHub client struct
type GitHub struct {
	token      string
	pageSize   int
	cache      *cache.Redis
	RateLimits prometheus.Counter
}

// New github client
func New(config config.Config, cache *cache.Redis) *GitHub {
	return &GitHub{
		token:    config.GitHubToken,
		pageSize: config.GitHubPageSize,
		cache:    cache,
		RateLimits: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: "starcharts",
			Subsystem: "github",
			Name:      "rate_limit_hits_total",
		}),
	}
}
