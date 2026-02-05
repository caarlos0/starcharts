package github

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/charmbracelet/log"
)

// Repository details.
type Repository struct {
	FullName        string `json:"full_name"`
	StargazersCount int    `json:"stargazers_count"`
	CreatedAt       string `json:"created_at"`
}

var ErrorNotFound = errors.New("Repository not found")

// RepoDetails gets the given repository details.
func (gh *GitHub) RepoDetails(ctx context.Context, name string) (Repository, error) {
	var repo Repository
	log := log.With("repo", name)

	var etag string
	etagKey := name + "_etag"

	if err := gh.cache.Get(etagKey, &etag); err != nil {
		log.Warn("failed to get etag from cache", "key", etagKey, "err", err)
	}

	resp, err := gh.makeRepoRequest(ctx, name, etag)
	if err != nil {
		return repo, err
	}

	bts, err := io.ReadAll(resp.Body)
	if err != nil {
		return repo, err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusNotModified:
		log.Info("not modified")
		effectiveEtags.Inc()
		err := gh.cache.Get(name, &repo)
		if err != nil {
			log.Warn("failed to get repo from cache", "name", name, "err", err)
			if err := gh.cache.Delete(etagKey); err != nil {
				log.Warn("failed to delete etag from cache", "key", etagKey, "err", err)
			}
			return gh.RepoDetails(ctx, name)
		}
		return repo, err
	case http.StatusForbidden:
		rateLimits.Inc()
		log.Warn("rate limit hit")
		return repo, ErrRateLimit
	case http.StatusOK:
		if err := json.Unmarshal(bts, &repo); err != nil {
			return repo, err
		}
		if err := gh.cache.Put(name, repo); err != nil {
			log.Warn("failed to cache repo", "name", name, "err", err)
		}

		etag = resp.Header.Get("etag")
		if etag != "" {
			if err := gh.cache.Put(etagKey, etag); err != nil {
				log.Warn("failed to cache etag", "key", etagKey, "err", err)
			}
		}

		return repo, nil
	case http.StatusNotFound:
		return repo, ErrorNotFound
	default:
		return repo, fmt.Errorf("%w: %v", ErrGitHubAPI, string(bts))
	}
}

func (gh *GitHub) makeRepoRequest(ctx context.Context, name, etag string) (*http.Response, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s", name)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	if etag != "" {
		req.Header.Add("If-None-Match", etag)
	}

	return gh.authorizedDo(req, 0)
}
