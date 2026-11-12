package controller

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/alicebob/miniredis"
	"github.com/caarlos0/starcharts/config"
	"github.com/caarlos0/starcharts/internal/cache"
	"github.com/caarlos0/starcharts/internal/github"
	"github.com/go-redis/redis"
	"github.com/gorilla/mux"
	"gopkg.in/h2non/gock.v1"
)

func TestGetRepoChart_PublicCache(t *testing.T) {
	defer gock.Off()
	gock.New("https://api.github.com").
		Get("/repos/test/public").
		Reply(http.StatusNotFound)

	mr, err := miniredis.Run()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(mr.Close)
	rc := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	c := cache.New(rc)
	t.Cleanup(func() { _ = c.Close() })
	if err := c.Put("test/public/[][][][]", "public-chart"); err != nil {
		t.Fatal(err)
	}

	r := mux.SetURLVars(httptest.NewRequest(http.MethodGet, "/test/public.svg", nil), map[string]string{
		"owner": "test",
		"repo":  "public",
	})
	w := httptest.NewRecorder()
	GetRepoChart(github.New(config.Config{}, c), c).ServeHTTP(w, r)
	if w.Code != http.StatusOK || w.Body.String() != "public-chart" {
		t.Fatalf("cached public chart: got status %d, body %q; want status %d, body %q",
			w.Code, w.Body.String(), http.StatusOK, "public-chart")
	}
}

func TestGetRepoChart_Private(t *testing.T) {
	defer gock.Off()
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(mr.Close)
	rc := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	c := cache.New(rc)
	t.Cleanup(func() { _ = c.Close() })
	gock.New("https://api.github.com").
		Get("/repos/test/private").
		Reply(200).
		JSON(map[string]any{"full_name": "test/private", "private": true, "visibility": "private"})

	r := mux.SetURLVars(httptest.NewRequest(http.MethodGet, "/test/private.svg", nil), map[string]string{
		"owner": "test",
		"repo":  "private",
	})
	w := httptest.NewRecorder()
	GetRepoChart(github.New(config.Config{}, c), c).ServeHTTP(w, r)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("private repository: got status %d, want %d", w.Code, http.StatusBadRequest)
	}
}
