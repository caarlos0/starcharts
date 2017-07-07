package github

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"sort"
	"sync"
	"time"

	"github.com/apex/log"
	cache "github.com/patrickmn/go-cache"
	"golang.org/x/sync/errgroup"
)

var (
	errNoMorePages = errors.New("no more pages to get")
	pageCache      *cache.Cache
)

const pageSize = 100

func init() {
	pageCache = cache.New(1*time.Hour, 2*time.Hour)
}

type Stargazer struct {
	StarredAt time.Time `json:"starred_at"`
}

func Stargazers(token string, repo Repository) (stars []Stargazer, err error) {
	sem := make(chan bool, 10)
	var g errgroup.Group
	var lock sync.Mutex
	for page := 0; page <= repo.StargazersCount/pageSize; page++ {
		sem <- true
		page := page
		g.Go(func() error {
			defer func() { <-sem }()
			result, err := getStargazersPage(token, repo.FullName, page)
			if err == errNoMorePages {
				return nil
			}
			if err != nil {
				return err
			}
			lock.Lock()
			defer lock.Unlock()
			stars = append(stars, result...)
			return nil
		})
	}
	err = g.Wait()
	sort.Slice(stars, func(i, j int) bool {
		return stars[i].StarredAt.Before(stars[j].StarredAt)
	})
	return
}

func getStargazersPage(token, name string, page int) (stars []Stargazer, err error) {
	var ctx = log.WithField("repo", name).WithField("page", page)
	cached, found := pageCache.Get(fmt.Sprintf("%s_%d", name, page))
	if found {
		ctx.Info("got from cache")
		return cached.([]Stargazer), nil
	}
	ctx.Infof("getting page from api")
	var url = fmt.Sprintf(
		"https://api.github.com/repos/%s/stargazers?page=%d&per_page=%d",
		name, page, pageSize,
	)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return stars, err
	}
	req.Header.Add("Accept", "application/vnd.github.v3.star+json")
	if token != "" {
		req.Header.Add("Authorization", "token "+token)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return stars, err
	}
	defer resp.Body.Close()

	// rate limit
	if resp.StatusCode == http.StatusForbidden {
		ctx.Warn("rate limit hit, waiting 5s before trying again")
		time.Sleep(5 * time.Second)
		return getStargazersPage(token, name, page)
	}
	if resp.StatusCode != http.StatusOK {
		bts, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return stars, err
		}
		return stars, fmt.Errorf("failed to get stargazers from github api: %v", string(bts))
	}
	err = json.NewDecoder(resp.Body).Decode(&stars)
	if len(stars) == 0 {
		return stars, errNoMorePages
	}
	pageCache.Set(fmt.Sprintf("%s_%d", name, page), stars, cache.DefaultExpiration)
	return
}
