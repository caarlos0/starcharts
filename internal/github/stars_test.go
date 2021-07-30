package github

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis"
	"github.com/caarlos0/starcharts/config"
	"github.com/caarlos0/starcharts/internal/cache"
	"github.com/caarlos0/starcharts/internal/roundrobin"
	"github.com/go-redis/redis"
	"github.com/matryer/is"
	"gopkg.in/h2non/gock.v1"
)

func TestStargazers(t *testing.T) {
	defer gock.Off()

	gock.New("https://api.github.com").
		Get("/rate_limit").
		Reply(200).
		JSON(rateLimit{rate{Limit: 5000, Remaining: 4000}})

	stargazers := []Stargazer{
		{StarredAt: time.Now()},
		{StarredAt: time.Now()},
	}

	repo := Repository{
		FullName:        "test/test",
		CreatedAt:       "2008-02-28T20:40:04Z",
		StargazersCount: 2,
	}

	mr, _ := miniredis.Run()
	rc := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	config := config.Get()
	cache := cache.New(rc)
	defer cache.Close()
	gt := New(config, cache)

	t.Run("get stargazers from api", func(t *testing.T) {
		is := is.New(t)
		gock.New("https://api.github.com").
			Get("/repos/test/test/stargazers").
			Reply(200).
			JSON(stargazers)
		_, err := gt.Stargazers(context.TODO(), repo)
		is.NoErr(err) // should not have errored
	})

	t.Run("get stargazers from cache", func(t *testing.T) {
		is := is.New(t)
		is.NoErr(cache.Put(repo.FullName+"_1_etag", "asdasd"))
		gock.New("https://api.github.com").
			Get("/repos/test/test/stargazers").
			MatchHeader("If-None-Match", "asdasd").
			Reply(304).
			JSON([]Stargazer{})
		_, err := gt.Stargazers(context.TODO(), repo)
		is.NoErr(err) // should not have errored
	})
}

func TestStargazers_EmptyResponseOnPagination(t *testing.T) {
	defer gock.Off()

	gock.New("https://api.github.com").
		Get("/rate_limit").
		Reply(200).
		JSON(rateLimit{rate{Limit: 5000, Remaining: 4000}})

	gock.New("https://api.github.com").
		Get("/rate_limit").
		Reply(200).
		JSON(rateLimit{rate{Limit: 5000, Remaining: 3999}})

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

	config := config.Get()
	cache := cache.New(rc)
	defer cache.Close()
	gt := New(config, cache)
	gt.pageSize = 2
	gt.tokens = roundrobin.New([]string{"12345"})

	t.Run("get stargazers from api", func(t *testing.T) {
		is := is.New(t)
		_, err := gt.Stargazers(context.TODO(), repo)
		is.NoErr(err) // should not have errored
	})
}

func TestStargazers_APIFailure(t *testing.T) {
	defer gock.Off()

	gock.New("https://api.github.com").
		Get("/rate_limit").
		Reply(200).
		JSON(rateLimit{rate{Limit: 5000, Remaining: 4000}})

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

	config := config.Get()
	cache := cache.New(rc)
	defer cache.Close()
	gt := New(config, cache)

	t.Run("set error if api return 404", func(t *testing.T) {
		is := is.New(t)
		_, err := gt.Stargazers(context.TODO(), repo1)
		is.True(err != nil) // should not have errored
	})
	t.Run("set error if api return 403", func(t *testing.T) {
		is := is.New(t)
		_, err := gt.Stargazers(context.TODO(), repo2)
		is.True(err != nil) // should not have errored
	})
}
