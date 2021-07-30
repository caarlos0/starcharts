package github

import (
	"context"
	"testing"

	"github.com/alicebob/miniredis"
	"github.com/caarlos0/starcharts/config"
	"github.com/caarlos0/starcharts/internal/cache"
	"github.com/caarlos0/starcharts/internal/roundrobin"
	"github.com/go-redis/redis"
	"github.com/matryer/is"
	"gopkg.in/h2non/gock.v1"
)

func TestRepoDetails(t *testing.T) {
	defer gock.Off()

	repo := Repository{
		FullName:        "test/test",
		CreatedAt:       "2008-02-28T20:40:04Z",
		StargazersCount: 3811,
	}

	mr, _ := miniredis.Run()
	rc := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	config := config.Get()
	cache := cache.New(rc)
	defer cache.Close()
	gt := New(config, cache)

	gock.New("https://api.github.com").
		Get("/rate_limit").
		Reply(200).
		JSON(rateLimit{rate{Limit: 5000, Remaining: 4000}})

	t.Run("get repo details from api", func(t *testing.T) {
		is := is.New(t)
		gock.New("https://api.github.com").
			Get("/repos/test/test").
			Reply(200).
			JSON(repo)
		_, err := gt.RepoDetails(context.TODO(), "test/test")
		is.NoErr(err) // should not fail to get from api
	})

	t.Run("get repo details from cache", func(t *testing.T) {
		is := is.New(t)
		is.NoErr(cache.Put("test/test_etag", "a"))
		gock.New("https://api.github.com").
			Get("/repos/test/test").
			MatchHeader("If-None-Match", "a").
			Reply(200).
			JSON(repo)
		_, err := gt.RepoDetails(context.TODO(), "test/test")
		is.NoErr(err) // should not fail to get from cache
	})
}

func TestRepoDetails_APIfailure(t *testing.T) {
	defer gock.Off()

	gock.New("https://api.github.com").
		Get("/rate_limit").
		Reply(200).
		JSON(rateLimit{rate{Limit: 5000, Remaining: 4000}})

	gock.New("https://api.github.com").
		Get("/repos/test/test").
		Reply(404)

	gock.New("https://api.github.com").
		Get("/repos/private/private").
		Reply(403)

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
		_, err := gt.RepoDetails(context.TODO(), "test/test")
		is.True(err != nil) // Expected error
	})
	t.Run("set error if api return 403", func(t *testing.T) {
		is := is.New(t)
		_, err := gt.RepoDetails(context.TODO(), "private/private")
		is.True(err != nil) // Expected error
	})
}

func TestRepoDetails_WithAuthToken(t *testing.T) {
	defer gock.Off()

	gock.New("https://api.github.com").
		Get("/rate_limit").
		Reply(200).
		JSON(rateLimit{rate{Limit: 5000, Remaining: 4000}})

	repo := Repository{
		FullName:        "aasm/aasm",
		CreatedAt:       "2008-02-28T20:40:04Z",
		StargazersCount: 3811,
	}

	gock.New("https://api.github.com").
		Get("/repos/test/private").
		Reply(200).
		JSON(repo)

	mr, _ := miniredis.Run()
	rc := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	config := config.Get()
	cache := cache.New(rc)
	defer cache.Close()
	gt := New(config, cache)
	gt.tokens = roundrobin.New([]string{"12345"})

	t.Run("get repo with auth token", func(t *testing.T) {
		is := is.New(t)
		_, err := gt.RepoDetails(context.TODO(), "test/private")
		is.NoErr(err) // should not fail to get from api with auth token
	})
}
