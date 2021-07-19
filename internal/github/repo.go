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
	err := gh.cache.Get(name, &repo)
	if err == nil {
		log.Info("got from cache")
		return repo, err
	}
	url := fmt.Sprintf("https://api.github.com/repos/%s", name)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return repo, err
	}
	if gh.token != "" {
		req.Header.Add("Authorization", fmt.Sprintf("token %s", gh.token))
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return repo, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusForbidden {
		gh.RateLimits.Inc()
		log.Warn("rate limit hit")
		return repo, ErrRateLimit
	}
	if resp.StatusCode != http.StatusOK {
		bts, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return repo, err
		}
		return repo, fmt.Errorf("%w: %v", ErrGitHubAPI, string(bts))
	}
	err = json.NewDecoder(resp.Body).Decode(&repo)
	if err := gh.cache.Put(name, repo); err != nil {
		log.Warn("failed to cache")
	}
	return repo, err
}
