package github

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"sort"
	"sync"
	"time"

	"github.com/apex/log"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
)

var errNoMorePages = errors.New("no more pages to get")

// Stargazer is a star at a given time
type Stargazer struct {
	StarredAt time.Time `json:"starred_at"`
}

// Stargazers returns all the stargazers of a given repo
func (gh *GitHub) Stargazers(repo Repository) (stars []Stargazer, err error) {
	sem := make(chan bool, 10)
	var g errgroup.Group
	var lock sync.Mutex
	for page := 1; page <= gh.lastPage(repo); page++ {
		sem <- true
		page := page
		g.Go(func() error {
			defer func() { <-sem }()
			result, err := gh.getStargazersPage(repo, page)
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

func (gh *GitHub) getStargazersPage(repo Repository, page int) (stars []Stargazer, err error) {
	var ctx = log.WithField("repo", repo.FullName).WithField("page", page)
	err = gh.cache.Get(fmt.Sprintf("%s_%d", repo.FullName, page), &stars)
	if err == nil {
		ctx.Info("got from cache")
		return
	}
	defer ctx.Trace("got page from api").Stop(&err)
	ctx.Infof("getting page from api")
	var url = fmt.Sprintf(
		"https://api.github.com/repos/%s/stargazers?page=%d&per_page=%d",
		repo.FullName,
		page,
		gh.pageSize,
	)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err == errNoMorePages {
		return stars, nil
	}
	if err != nil {
		return stars, err
	}
	req.Header.Add("Accept", "application/vnd.github.v3.star+json")
	if gh.token != "" {
		req.Header.Add("Authorization", fmt.Sprintf("token %s", gh.token))
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return stars, err
	}
	defer resp.Body.Close()

	// rate limit
	if resp.StatusCode == http.StatusForbidden {
		gh.RateLimits.Inc()
		ctx.Warn("rate limit hit")
		return stars, errors.Wrap(err, "rate limited, please try again later")
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
	var expire = time.Hour * 24 * 7
	if page == gh.lastPage(repo) {
		expire = time.Hour * 2
	}
	ctx.WithField("expire", expire).Info("caching...")
	if err := gh.cache.Put(
		fmt.Sprintf("%s_%d", repo.FullName, page),
		stars,
		expire,
	); err != nil {
		ctx.WithError(err).Warn("failed to cache")
	}
	return
}

func (gh *GitHub) lastPage(repo Repository) int {
	return (repo.StargazersCount / gh.pageSize) + 1
}
