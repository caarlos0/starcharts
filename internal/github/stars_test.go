package github

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/alicebob/miniredis"
	"github.com/caarlos0/starcharts/config"
	"github.com/caarlos0/starcharts/internal/cache"
	"github.com/go-redis/redis"
	"github.com/matryer/is"
	"gopkg.in/h2non/gock.v1"
)

// 2017-07-02T00:00:00Z, a Sunday.
const firstWeek = 1498953600

const week = 7 * 24 * time.Hour

func newTestGitHub(t *testing.T) *GitHub {
	t.Helper()

	mr, err := miniredis.Run()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(mr.Close)

	c := cache.New(redis.NewClient(&redis.Options{Addr: mr.Addr()}))
	t.Cleanup(func() { _ = c.Close() })

	return New(config.Get(), c)
}

func mockStarHistory(page int, link string, weeks []historyWeek) {
	gock.New("https://api.github.com").
		Get("/repos/test/test/stargazers/history").
		MatchParam("page", fmt.Sprintf("%d", page)).
		Reply(200).
		SetHeader("Link", link).
		SetHeader("etag", fmt.Sprintf(`"etag-%d"`, page)).
		JSON(weeks)
}

func TestStargazers(t *testing.T) {
	defer gock.Off()

	repo := Repository{
		FullName:        "test/test",
		CreatedAt:       "2017-07-07T01:08:03Z",
		StargazersCount: 10,
	}

	// the API returns the most recent week first.
	mockStarHistory(1, `<https://api.github.com/repositories/1/stargazers/history?page=2>; rel="next", `+
		`<https://api.github.com/repositories/1/stargazers/history?page=2>; rel="last"`, []historyWeek{
		{Week: firstWeek + 2*7*24*3600, Days: []int{0, 0, 0, 0, 0, 0, 0}},
		{Week: firstWeek + 7*24*3600, Days: []int{0, 1, 0, 0, 0, 2, 0}},
	})
	mockStarHistory(2, `<https://api.github.com/repositories/1/stargazers/history?page=1>; rel="prev"`, []historyWeek{
		{Week: firstWeek, Days: []int{4, 0, 0, 0, 0, 0, 3}},
	})

	gt := newTestGitHub(t)

	is := is.New(t)
	stars, err := gt.Stargazers(context.TODO(), repo)
	is.NoErr(err) // should not have errored

	start := time.Unix(firstWeek, 0).UTC()
	is.Equal([]Stargazer{
		{StarredAt: start, Count: 4},
		{StarredAt: start.AddDate(0, 0, 6), Count: 7},
		{StarredAt: start.Add(week).AddDate(0, 0, 1), Count: 8},
		{StarredAt: start.Add(week).AddDate(0, 0, 5), Count: 10},
	}, stars[:len(stars)-1]) // all but the trailing "now" data point

	last := stars[len(stars)-1]
	is.Equal(repo.StargazersCount, last.Count)           // should end on the current star count
	is.True(time.Since(last.StarredAt) < 10*time.Second) // should end at about now
	is.True(gock.IsDone())                               // should have consumed all mocks
}

func TestStargazersUsesEtagCache(t *testing.T) {
	defer gock.Off()

	repo := Repository{
		FullName:        "test/test",
		CreatedAt:       "2017-07-07T01:08:03Z",
		StargazersCount: 2,
	}

	weeks := []historyWeek{{Week: firstWeek, Days: []int{0, 0, 2, 0, 0, 0, 0}}}
	mockStarHistory(1, "", weeks)

	gt := newTestGitHub(t)

	is := is.New(t)
	first, err := gt.Stargazers(context.TODO(), repo)
	is.NoErr(err) // should not have errored

	gock.New("https://api.github.com").
		Get("/repos/test/test/stargazers/history").
		MatchHeader("If-None-Match", `"etag-1"`).
		Reply(304)

	second, err := gt.Stargazers(context.TODO(), repo)
	is.NoErr(err)                                          // should not have errored
	is.Equal(first[:len(first)-1], second[:len(second)-1]) // should serve the same data from cache
	is.True(gock.IsDone())                                 // should have consumed all mocks
}

func TestStargazersAPIFailure(t *testing.T) {
	for name, status := range map[string]int{
		"not found":  404,
		"rate limit": 403,
	} {
		t.Run(name, func(t *testing.T) {
			defer gock.Off()

			gock.New("https://api.github.com").
				Get("/repos/test/test/stargazers/history").
				Persist().
				Reply(status).
				JSON([]historyWeek{})

			gt := newTestGitHub(t)

			is := is.New(t)
			_, err := gt.Stargazers(context.TODO(), Repository{
				FullName:        "test/test",
				CreatedAt:       "2017-07-07T01:08:03Z",
				StargazersCount: 3,
			})
			is.True(err != nil) // should have errored
		})
	}
}

func TestToStargazers(t *testing.T) {
	t.Run("no stars", func(t *testing.T) {
		is := is.New(t)
		is.Equal(0, len(toStargazers([]historyWeek{
			{Week: firstWeek, Days: []int{0, 0, 0, 0, 0, 0, 0}},
		}))) // should have no data points
	})

	t.Run("sorts weeks chronologically", func(t *testing.T) {
		is := is.New(t)
		stars := toStargazers([]historyWeek{
			{Week: firstWeek + 7*24*3600, Days: []int{1, 0, 0, 0, 0, 0, 0}},
			{Week: firstWeek, Days: []int{0, 0, 0, 5, 0, 0, 0}},
		})
		is.Equal([]int{5, 6}, []int{stars[0].Count, stars[1].Count}) // should accumulate in order
	})
}

func TestParseLastPageFromLink(t *testing.T) {
	is := is.New(t)

	is.Equal(0, parseLastPageFromLink(""))
	is.Equal(0, parseLastPageFromLink(`<https://api.github.com/x?page=2>; rel="next"`))
	is.Equal(17, parseLastPageFromLink(
		`<https://api.github.com/repositories/1/stargazers/history?per_page=30&page=2>; rel="next", `+
			`<https://api.github.com/repositories/1/stargazers/history?per_page=30&page=17>; rel="last"`,
	))
}
