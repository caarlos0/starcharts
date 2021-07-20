package github

import (
	"errors"

	"github.com/caarlos0/starcharts/config"
	"github.com/caarlos0/starcharts/internal/cache"
	"github.com/prometheus/client_golang/prometheus"
)

// ErrRateLimit happens when we rate limit github API.
var ErrRateLimit = errors.New("rate limited, please try again later")

// ErrGitHubAPI happens when github responds with something other than a 2xx.
var ErrGitHubAPI = errors.New("failed to talk with github api")

// GitHub client struct.
type GitHub struct {
	token    string
	pageSize int
	cache    *cache.Redis

	rateLimits, effectiveEtags prometheus.Counter
}

var rateLimits = prometheus.NewCounter(prometheus.CounterOpts{
	Namespace: "starcharts",
	Subsystem: "github",
	Name:      "rate_limit_hits_total",
})

var effectiveEtags = prometheus.NewCounter(prometheus.CounterOpts{
	Namespace: "starcharts",
	Subsystem: "github",
	Name:      "effective_etag_uses_total",
})

func init() {
	prometheus.MustRegister(rateLimits, effectiveEtags)
}

// New github client.
func New(config config.Config, cache *cache.Redis) *GitHub {

	return &GitHub{
		token:          config.GitHubToken,
		pageSize:       config.GitHubPageSize,
		cache:          cache,
		rateLimits:     rateLimits,
		effectiveEtags: effectiveEtags,
	}
}
