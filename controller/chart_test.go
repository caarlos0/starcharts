package controller

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/alicebob/miniredis"
	"github.com/caarlos0/starcharts/config"
	"github.com/caarlos0/starcharts/internal/cache"
	"github.com/caarlos0/starcharts/internal/github"
	"github.com/go-redis/redis"
	"github.com/gorilla/mux"
	"gopkg.in/h2non/gock.v1"
)

func TestGetRepoChart_LegacyPrivateCache(t *testing.T) {
	defer gock.Off()
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(mr.Close)
	rc := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	c := cache.New(rc)
	t.Cleanup(func() { _ = c.Close() })
	if err := c.Put("test/private/[][][][]", "private-chart"); err != nil {
		t.Fatal(err)
	}
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
	if strings.Contains(w.Body.String(), "private-chart") {
		t.Fatal("served private chart cached before visibility checks")
	}
	if w.Code != http.StatusBadRequest {
		t.Fatalf("private repository: got status %d, want %d", w.Code, http.StatusBadRequest)
	}
}
