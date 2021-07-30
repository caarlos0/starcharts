package github

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/apex/log"
	"github.com/caarlos0/starcharts/config"
	"github.com/caarlos0/starcharts/internal/cache"
	"github.com/caarlos0/starcharts/internal/roundrobin"
	"github.com/prometheus/client_golang/prometheus"
)

// ErrRateLimit happens when we rate limit github API.
var ErrRateLimit = errors.New("rate limited, please try again later")

// ErrGitHubAPI happens when github responds with something other than a 2xx.
var ErrGitHubAPI = errors.New("failed to talk with github api")

// GitHub client struct.
type GitHub struct {
	tokens   roundrobin.RoundRobiner
	pageSize int
	cache    *cache.Redis

	rateLimits, effectiveEtags, retryNewTokens prometheus.Counter
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

var retryNewTokens = prometheus.NewCounter(prometheus.CounterOpts{
	Namespace: "starcharts",
	Subsystem: "github",
	Name:      "next_token_retries",
})

func init() {
	prometheus.MustRegister(rateLimits, effectiveEtags, retryNewTokens)
}

// New github client.
func New(config config.Config, cache *cache.Redis) *GitHub {
	return &GitHub{
		tokens:         roundrobin.New(config.GitHubTokens),
		pageSize:       config.GitHubPageSize,
		cache:          cache,
		rateLimits:     rateLimits,
		effectiveEtags: effectiveEtags,
		retryNewTokens: retryNewTokens,
	}
}

func (gh *GitHub) authorizedDo(req *http.Request) (*http.Response, error) {
	token, err := gh.tokens.Pick()
	if err != nil || token == nil {
		log.WithError(err).Error("couldn't get a valid token")
		return http.DefaultClient.Do(req)
	}

	req.Header.Add("Authorization", fmt.Sprintf("token %s", token.Key()))
	resp, err := http.DefaultClient.Do(req)
	if resp.StatusCode == http.StatusUnauthorized {
		token.Invalidate()
	}
	return resp, err
}
