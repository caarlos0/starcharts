package github

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis"
	"github.com/caarlos0/starcharts/config"
	"github.com/caarlos0/starcharts/internal/cache"
	"github.com/go-redis/redis"
	"gopkg.in/h2non/gock.v1"
)

func TestStargazers(t *testing.T) {
	defer gock.Off()

	stargazers := []Stargazer{
		{StarredAt: time.Now()},
		{StarredAt: time.Now()},
	}

	repo := Repository{
		FullName:        "test/test",
		CreatedAt:       "2008-02-28T20:40:04Z",
		StargazersCount: 2,
	}

	gock.New("https://api.github.com").
		Get("/repos/test/test/stargazers").
		Reply(200).
		JSON(stargazers)

	mr, _ := miniredis.Run()
	rc := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	var config = config.Get()
	var cache = cache.New(rc)
	defer cache.Close()
	var gt = New(config, cache)

	t.Run("get stargazers from api", func(t *testing.T) {
		_, err := gt.Stargazers(context.TODO(), repo)
		if err != nil {
			t.Errorf("RepoDetails returned error %v", err)
		}
	})

	t.Run("get stargazers from cache", func(t *testing.T) {
		_, err := gt.Stargazers(context.TODO(), repo)
		if err != nil {
			t.Errorf("RepoDetails returned error %v", err)
		}
	})
}

func TestStargazers_EmptyResponseOnPagination(t *testing.T) {
	defer gock.Off()

	stargazers := []Stargazer{
		{StarredAt: time.Now()},
		{StarredAt: time.Now()},
	}

	repo := Repository{
		FullName:        "test/test",
		CreatedAt:       "2008-02-28T20:40:04Z",
		StargazersCount: 3,
	}

	gock.New("https://api.github.com").
		Get("/repos/test/test/stargazers").
		MatchParam("page", "1").
		MatchParam("per_page", "2").
		Reply(200).
		JSON(stargazers)

	gock.New("https://api.github.com").
		Get("/repos/test/test/stargazers").
		MatchParam("page", "2").
		MatchParam("per_page", "2").
		Reply(200).
		JSON([]Stargazer{})

	mr, _ := miniredis.Run()
	rc := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	var config = config.Get()
	var cache = cache.New(rc)
	defer cache.Close()
	var gt = New(config, cache)
	gt.pageSize = 2
	gt.token = "12345"

	t.Run("get stargazers from api", func(t *testing.T) {
		_, err := gt.Stargazers(context.TODO(), repo)
		if err != nil {
			t.Errorf("RepoDetails returned error %v", err)
		}
	})
}

func TestStargazers_APIFailure(t *testing.T) {
	defer gock.Off()

	repo1 := Repository{
		FullName:        "test/test",
		CreatedAt:       "2008-02-28T20:40:04Z",
		StargazersCount: 3,
	}

	repo2 := Repository{
		FullName:        "private/private",
		CreatedAt:       "2008-02-28T20:40:04Z",
		StargazersCount: 3,
	}

	gock.New("https://api.github.com").
		Get("/repos/test/test/stargazers").
		Persist().
		Reply(404).
		JSON([]Stargazer{})

	gock.New("https://api.github.com").
		Get("/repos/private/private/stargazers").
		Persist().
		Reply(403).
		JSON([]Stargazer{})

	mr, _ := miniredis.Run()
	rc := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	var config = config.Get()
	var cache = cache.New(rc)
	defer cache.Close()
	var gt = New(config, cache)

	t.Run("set error if api return 404", func(t *testing.T) {
		details, err := gt.Stargazers(context.TODO(), repo1)
		if err == nil {
			t.Errorf("Expected error but got %v", details)
		}
	})
	t.Run("set error if api return 403", func(t *testing.T) {
		details, err := gt.Stargazers(context.TODO(), repo2)
		if err == nil {
			t.Errorf("Expected error but got %v", details)
		}
	})
}
