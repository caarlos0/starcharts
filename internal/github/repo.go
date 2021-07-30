package github

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/apex/log"
)

// Repository details.
type Repository struct {
	FullName        string `json:"full_name"`
	StargazersCount int    `json:"stargazers_count"`
	CreatedAt       string `json:"created_at"`
}

// RepoDetails gets the given repository details.
func (gh *GitHub) RepoDetails(ctx context.Context, name string) (Repository, error) {
	var repo Repository
	log := log.WithField("repo", name)

	var etag string
	etagKey := name + "_etag"

	if err := gh.cache.Get(etagKey, &etag); err != nil {
		log.WithError(err).Warnf("failed to get %s from cache", etagKey)
	}

	resp, err := gh.makeRepoRequest(ctx, name, etag)
	if err != nil {
		return repo, err
	}

	bts, err := ioutil.ReadAll(resp.Body)
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
			log.WithError(err).Warnf("failed to get %s from cache", name)
			if err := gh.cache.Delete(etagKey); err != nil {
				log.WithError(err).Warnf("failed to delete %s from cache", etagKey)
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
			log.WithError(err).Warnf("failed to cache %s", name)
		}

		etag = resp.Header.Get("etag")
		if etag != "" {
			if err := gh.cache.Put(etagKey, etag); err != nil {
				log.WithError(err).Warnf("failed to cache %s", etagKey)
			}
		}

		return repo, nil
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
