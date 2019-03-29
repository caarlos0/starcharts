package github

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"time"

	"github.com/apex/log"
)

// Repository details
type Repository struct {
	FullName        string `json:"full_name"`
	StargazersCount int    `json:"stargazers_count"`
	CreatedAt       string `json:"created_at"`
}

// RepoDetails gets the given repository details
func (gh *GitHub) RepoDetails(name string) (repo Repository, err error) {
	var ctx = log.WithField("repo", name)
	err = gh.cache.Get(name, &repo)
	if err == nil {
		ctx.Info("got from cache")
		return
	}
	var url = fmt.Sprintf("https://api.github.com/repos/%s", name)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return
	}
	if gh.token != "" {
		req.Header.Add("Authorization", fmt.Sprintf("token %s", gh.token))
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusForbidden {
		ctx.Warn("rate limit hit, waiting 10s before trying again")
		time.Sleep(10 * time.Second)
		return gh.RepoDetails(name)
	}
	if resp.StatusCode != http.StatusOK {
		bts, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return repo, err
		}
		return repo, fmt.Errorf("failed to get stargazers from github api: %v", string(bts))
	}
	err = json.NewDecoder(resp.Body).Decode(&repo)
	if err := gh.cache.Put(name, repo, time.Hour*2); err != nil {
		ctx.Warn("failed to cache")
	}
	return
}
