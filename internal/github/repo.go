package github

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	cache "github.com/patrickmn/go-cache"
)

var repoCache *cache.Cache

func init() {
	repoCache = cache.New(1*time.Hour, 2*time.Hour)
}

type Repository struct {
	FullName        string `json:"full_name"`
	StargazersCount int    `json:"stargazers_count"`
}

func RepoDetails(token, name string) (repo Repository, err error) {
	cached, found := repoCache.Get(name)
	if found {
		return cached.(Repository), nil
	}

	var url = fmt.Sprintf("https://api.github.com/repos/%s", name)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return
	}
	if token != "" {
		req.Header.Add("Authorization", "token "+token)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	err = json.NewDecoder(resp.Body).Decode(&repo)
	repoCache.Set(name, repo, cache.DefaultExpiration)
	return
}
