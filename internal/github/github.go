package github

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/apex/log"
	"github.com/caarlos0/starcharts/config"
	"github.com/caarlos0/starcharts/internal/cache"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/shurcooL/githubv4"
	"golang.org/x/oauth2"
)

// GitHub client struct.
type GitHub struct {
	client     *githubv4.Client
	cache      *cache.Redis
	RateLimits prometheus.Counter
}

// New github client.
func New(config config.Config, cache *cache.Redis) *GitHub {
	src := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: config.GitHubToken},
	)
	httpClient := oauth2.NewClient(context.Background(), src)
	client := githubv4.NewClient(httpClient)

	return &GitHub{
		client: client,
		cache:  cache,
		RateLimits: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: "starcharts",
			Subsystem: "github",
			Name:      "rate_limit_hits_total",
		}),
	}
}

// Stargazer is a star at a given time.
type Stargazer struct {
	StarredAt time.Time
}

// FullRepoInfo contains information about the repo and its stars.
type FullRepoInfo struct {
	FullName        string
	StargazersCount int32
	CreatedAt       string
}

// Repository details.
type Repository struct {
	FullName        string
	StargazersCount int
	CreatedAt       string
}

var cacheExpire = time.Hour * 24 * 7

func (gh *GitHub) RepoDetails(ctx context.Context, owner, name string) (Repository, error) {
	var cold Repository
	key := fmt.Sprintf("r:%s/%s", owner, name)
	if err := gh.cache.Get(key, &cold); err == nil {
		log.Info("got from cache")
		return cold, nil
	}

	hot, err := gh.doGetRepoDetails(ctx, owner, name)
	if err != nil {
		return hot, err
	}

	log.Info("caching...")
	if err := gh.cache.Put(
		key,
		hot,
		cacheExpire,
	); err != nil {
		log.WithError(err).Warn("failed to cache")
	}

	return hot, err
}

func (gh *GitHub) Stargazers(ctx context.Context, owner, name string) ([]Stargazer, error) {
	var cold []Stargazer
	key := fmt.Sprintf("s:%s/%s", owner, name)
	if err := gh.cache.Get(key, &cold); err == nil {
		log.Info("got from cache")
		return cold, nil
	}

	hot, err := gh.doGetStargazers(ctx, owner, name)
	if err != nil {
		return hot, err
	}

	log.Info("caching...")
	if err := gh.cache.Put(
		key,
		hot,
		cacheExpire,
	); err != nil {
		log.WithError(err).Warn("failed to cache")
	}

	return hot, err
}

func (gh *GitHub) doGetRepoDetails(ctx context.Context, owner string, name string) (Repository, error) {
	var query struct {
		Repository struct {
			NameWithOwner githubv4.String
			CreatedAt     githubv4.DateTime
			Stargazers    struct {
				TotalCount githubv4.Int
			} `graphql:"stargazers(first: 0)"`
		} `graphql:"repository(owner: $owner, name: $name)"`
	}

	variables := map[string]interface{}{
		"owner": githubv4.String(owner),
		"name":  githubv4.String(name),
	}

	if err := gh.queryWithRetry(ctx, &query, variables); err != nil {
		return Repository{}, err
	}

	log.WithField("repo", query.Repository.NameWithOwner).Info("got")

	return Repository{
		FullName:        string(query.Repository.NameWithOwner),
		StargazersCount: int(query.Repository.Stargazers.TotalCount),
		CreatedAt:       query.Repository.CreatedAt.Time.Format(time.RFC3339),
	}, nil
}

func (gh *GitHub) doGetStargazers(ctx context.Context, owner, name string) ([]Stargazer, error) {
	var query struct {
		Repository struct {
			Stargazers struct {
				Edges []struct {
					StarredAt githubv4.DateTime
					Cursor    githubv4.String
				}
			} `graphql:"stargazers(first: 100, after:$cursor)"`
		} `graphql:"repository(owner: $owner, name: $name)"`
	}

	variables := map[string]interface{}{
		"owner":  githubv4.String(owner),
		"name":   githubv4.String(name),
		"cursor": (*githubv4.String)(nil),
	}

	var stargazers []Stargazer
	for {
		log.Debugf("cursor is %v", variables["cursor"])
		err := gh.queryWithRetry(ctx, &query, variables)
		if err != nil {
			return stargazers, err
		}

		if len(query.Repository.Stargazers.Edges) == 0 {
			break
		}

		for _, v := range query.Repository.Stargazers.Edges {
			stargazers = append(stargazers, Stargazer{
				StarredAt: v.StarredAt.Time,
			})

			variables["cursor"] = &v.Cursor
		}
	}

	return stargazers, nil
}

func (gh *GitHub) queryWithRetry(ctx context.Context, q interface{}, variables map[string]interface{}) error {
	if err := gh.client.Query(ctx, q, variables); err != nil {
		if strings.Contains(err.Error(), "abuse-rate-limits") {
			gh.RateLimits.Inc()
			time.Sleep(time.Minute)
			return gh.queryWithRetry(ctx, q, variables)
		}
		return err
	}

	return nil
}
