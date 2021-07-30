package github

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
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

var rateLimiters = prometheus.NewGaugeVec(prometheus.GaugeOpts{
	Namespace: "starcharts",
	Subsystem: "github",
	Name:      "rate_limit_remaining",
}, []string{"token"})


func init() {
	prometheus.MustRegister(rateLimits, effectiveEtags, retryNewTokens,rateLimiters)
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

const maxTries = 3

func (gh *GitHub) authorizedDo(req *http.Request, try int) (*http.Response, error) {
	if try > maxTries {
		return nil, fmt.Errorf("couldn't find a valid token")
	}
	token, err := gh.tokens.Pick()
	if err != nil || token == nil {
		log.WithError(err).Error("couldn't get a valid token")
		return http.DefaultClient.Do(req)
	}

	ok, err := gh.checkRateLimit(token)
	if err != nil {
		log.WithError(err).Error("couldn't check rate limit, trying next token")
		return gh.authorizedDo(req, try+1)
	}
	if !ok {
		log.Warn("skipping token because it used too much of its limit")
		return gh.authorizedDo(req, try+1)
	}

	req.Header.Add("Authorization", fmt.Sprintf("token %s", token.Key()))
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return resp, err
	}
	if resp.StatusCode == http.StatusUnauthorized {
		token.Invalidate()
	}
	return resp, err
}

func (gh *GitHub) checkRateLimit(token *roundrobin.Token) (bool, error) {
	req, err := http.NewRequest(http.MethodGet, "https://api.github.com/rate_limit", nil)
	if err != nil {
		return false, err
	}
	req.Header.Add("Authorization", fmt.Sprintf("token %s", token.Key()))
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return false, err
	}

	bts, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, err
	}

	var limit rateLimit
	if err := json.Unmarshal(bts, &limit); err != nil {
		return false, err
	}
	rate := limit.Rate
	log.Debugf("%s rate %d/%d", token, rate.Remaining, rate.Limit)
	rateLimiters.WithLabelValues(token.String()).Set(float64(rate.Remaining))
	return rate.Remaining > rate.Limit/2, nil // allow at most 50% rate limit usage
}

type rateLimit struct {
	Rate struct {
		Remaining int `json:"remaining"`
		Limit     int `json:"limit"`
	} `json:"rate"`
}
