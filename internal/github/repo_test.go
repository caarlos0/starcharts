package github

import (
	"context"
	"testing"

	"github.com/alicebob/miniredis"
	"github.com/caarlos0/starcharts/config"
	"github.com/caarlos0/starcharts/internal/cache"
	"github.com/go-redis/redis"
	"gopkg.in/h2non/gock.v1"
)

func TestRepoDetails(t *testing.T) {
	defer gock.Off()

	repo := Repository{
		FullName:        "test/test",
		CreatedAt:       "2008-02-28T20:40:04Z",
		StargazersCount: 3811,
	}

	gock.New("https://api.github.com").
		Get("/repos/test/test").
		Reply(200).
		JSON(repo)

	mr, _ := miniredis.Run()
	rc := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	var config = config.Get()
	var cache = cache.New(rc)
	defer cache.Close()
	var gt = New(config, cache)

	t.Run("get repo details from api", func(t *testing.T) {
		_, err := gt.RepoDetails(context.TODO(), "test/test")
		if err != nil {
			t.Errorf("RepoDetails returned error %v", err)
		}
	})

	t.Run("get repo details from cache", func(t *testing.T) {
		_, err := gt.RepoDetails(context.TODO(), "test/test")
		if err != nil {
			t.Errorf("RepoDetails returned error %v", err)
		}
	})
}

func TestRepoDetails_APIfailure(t *testing.T) {
	defer gock.Off()

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

	var config = config.Get()
	var cache = cache.New(rc)
	defer cache.Close()
	var gt = New(config, cache)

	t.Run("set error if api return 404", func(t *testing.T) {
		details, err := gt.RepoDetails(context.TODO(), "test/test")
		if err == nil {
			t.Errorf("Expected error but got %v", details)
		}
	})
	t.Run("set error if api return 403", func(t *testing.T) {
		details, err := gt.RepoDetails(context.TODO(), "private/private")
		if err == nil {
			t.Errorf("Expected error but got %v", details)
		}
	})
}

func TestRepoDetails_WithAuthToken(t *testing.T) {
	defer gock.Off()

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

	var config = config.Get()
	var cache = cache.New(rc)
	defer cache.Close()
	var gt = New(config, cache)
	gt.token = "12345"

	t.Run("get repo with auth token", func(t *testing.T) {
		_, err := gt.RepoDetails(context.TODO(), "test/private")
		if err != nil {
			t.Errorf("RepoDetails returned error %v", err)
		}
	})
}
