package github

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/apex/log"
)

// Repository details.
type Repository struct {
	FullName        string `json:"full_name"`
	StargazersCount int    `json:"stargazers_count"`
	CreatedAt       string `json:"created_at"`
}

// RepoDetails gets the given repository details.
func (gh *GitHub) RepoDetails(name string) (Repository, error) {
	var repo Repository
	var ctx = log.WithField("repo", name)
	err := gh.cache.Get(name, &repo)
	if err == nil {
		ctx.Info("got from cache")
		return repo, err
	}
	var url = fmt.Sprintf("https://api.github.com/repos/%s", name)
	req, err := http.NewRequest(http.MethodGet, url, nil)
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
		ctx.Warn("rate limit hit")
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
	if err := gh.cache.Put(name, repo, time.Hour*2); err != nil {
		ctx.Warn("failed to cache")
	}
	return repo, err
}
