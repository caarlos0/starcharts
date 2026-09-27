package github

import (
	"cmp"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"regexp"
	"slices"
	"strconv"
	"time"

	"golang.org/x/sync/errgroup"
)

// linkLastPageRegex is used to parse the last page number from the Link header.
var linkLastPageRegex = regexp.MustCompile(`[&?]page=(\d+)[^>]*>;\s*rel="last"`)

const (
	// maxConcurrentRequests is the maximum number of concurrent requests to GitHub API.
	maxConcurrentRequests = 5
	// historyPageSize is the maximum per_page the star history endpoint accepts.
	historyPageSize = 30
)

// Stargazer is the total number of stars a repository had at a given time.
type Stargazer struct {
	StarredAt time.Time
	Count     int
}

// historyWeek is a single week of the repository star history.
//
// Week is the unix timestamp of the first day of the week, and Days holds the
// number of stars created on each day of that week, starting on Sunday.
type historyWeek struct {
	Week int64 `json:"week"`
	Days []int `json:"days"`
}

// historyPage is a page of the star history, alongside the last page number
// advertised by the API, so it survives a cache round trip.
type historyPage struct {
	Weeks    []historyWeek
	LastPage int
}

// Stargazers returns the star count over time of a given repo.
func (gh *GitHub) Stargazers(ctx context.Context, repo Repository) ([]Stargazer, error) {
	weeks, err := gh.starHistory(ctx, repo)
	if err != nil {
		return nil, err
	}
	return toStargazers(weeks), nil
}

// starHistory fetches every week of the repository star history.
func (gh *GitHub) starHistory(ctx context.Context, repo Repository) ([]historyWeek, error) {
	// the last page is only known after the first request, as it comes from
	// the Link header.
	first, err := gh.getHistoryPage(ctx, repo, 1)
	if err != nil {
		return nil, err
	}

	slog.Debug(
		"got star history pagination info",
		"repo", repo.FullName,
		"lastPage", first.LastPage,
		"starCount", repo.StargazersCount,
	)

	if first.LastPage <= 1 {
		return first.Weeks, nil
	}

	// pages are collected into their own slot, so no locking is needed and
	// the result stays in page order.
	pages := make([][]historyWeek, first.LastPage)
	pages[0] = first.Weeks

	var group errgroup.Group
	group.SetLimit(maxConcurrentRequests)
	for page := 2; page <= first.LastPage; page++ {
		group.Go(func() error {
			result, err := gh.getHistoryPage(ctx, repo, page)
			if err != nil {
				return err
			}
			pages[page-1] = result.Weeks
			return nil
		})
	}
	if err := group.Wait(); err != nil {
		return nil, err
	}
	return slices.Concat(pages...), nil
}

// toStargazers turns the weekly star history into a cumulative daily series.
//
// Days in which no star was given are left out: the chart interpolates between
// data points, so they would only add noise.
func toStargazers(weeks []historyWeek) []Stargazer {
	// the API returns the most recent week first.
	slices.SortFunc(weeks, func(a, b historyWeek) int {
		return cmp.Compare(a.Week, b.Week)
	})

	var stars []Stargazer
	count := 0
	for _, week := range weeks {
		start := time.Unix(week.Week, 0).UTC()
		for day, added := range week.Days {
			if added == 0 {
				continue
			}
			count += added
			stars = append(stars, Stargazer{
				StarredAt: start.AddDate(0, 0, day),
				Count:     count,
			})
		}
	}

	if count == 0 {
		return nil
	}

	// extend the chart up to now, as the last star might have been given a
	// while ago.
	return append(stars, Stargazer{
		StarredAt: time.Now().UTC(),
		Count:     count,
	})
}

// - get last modified from cache
//   - if exists, hit api with it
//     - if it returns 304, get from cache
//       - if succeeds, return it
//       - if fails, it means we dont have that page in cache, hit api again
//         - if succeeds, cache and return both the api and header
//         - if fails, return error
//   - if not exists, hit api
//     - if succeeds, cache and return both the api and header
//     - if fails, return error

// nolint: funlen
// TODO: refactor.
func (gh *GitHub) getHistoryPage(ctx context.Context, repo Repository, page int) (historyPage, error) {
	log := slog.With("repo", repo.FullName, "page", page)
	start := time.Now()
	defer func() {
		log.Debug("get page", "duration", time.Since(start))
	}()

	var result historyPage
	key := fmt.Sprintf("%s_history_%d", repo.FullName, page)
	etagKey := key + "_etag"

	var etag string
	if err := gh.cache.Get(etagKey, &etag); err != nil {
		log.Warn("failed to get from cache", "etag", etagKey, "error", err)
	}

	resp, err := gh.makeStarHistoryRequest(ctx, repo, page, etag)
	if err != nil {
		return result, err
	}
	defer resp.Body.Close() //nolint:errcheck

	bts, err := io.ReadAll(resp.Body)
	if err != nil {
		return result, err
	}

	switch resp.StatusCode {
	case http.StatusNotModified:
		effectiveEtags.Inc()
		log.Info("not modified")
		if err := gh.cache.Get(key, &result); err != nil {
			log.Warn("failed to get from cache", "key", key, "error", err)
			if err := gh.cache.Delete(etagKey); err != nil {
				log.Warn("failed to delete from cache", "etag", etagKey, "error", err)
			}
			return gh.getHistoryPage(ctx, repo, page)
		}
		return result, nil
	case http.StatusForbidden:
		rateLimits.Inc()
		log.Warn("rate limit hit")
		return result, ErrRateLimit
	case http.StatusOK:
		if err := json.Unmarshal(bts, &result.Weeks); err != nil {
			return result, err
		}

		result.LastPage = max(parseLastPageFromLink(resp.Header.Get("Link")), 1)
		log.Debug("parsed last page from Link header", "lastPage", result.LastPage)

		if err := gh.cache.Put(key, result); err != nil {
			log.Warn("failed to cache", "key", key, "error", err)
		}

		if etag := resp.Header.Get("etag"); etag != "" {
			if err := gh.cache.Put(etagKey, etag); err != nil {
				log.Warn("failed to cache", "etag", etagKey, "error", err)
			}
		}

		return result, nil
	default:
		return result, fmt.Errorf("%w: %v", ErrGitHubAPI, string(bts))
	}
}

func (gh *GitHub) makeStarHistoryRequest(ctx context.Context, repo Repository, page int, etag string) (*http.Response, error) {
	url := fmt.Sprintf(
		"https://api.github.com/repos/%s/stargazers/history?page=%d&per_page=%d",
		repo.FullName,
		page,
		historyPageSize,
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Accept", "application/vnd.github+json")
	req.Header.Add("X-GitHub-Api-Version", "2026-03-10")
	if etag != "" {
		req.Header.Add("If-None-Match", etag)
	}

	return gh.authorizedDo(req, 0)
}

// parseLastPageFromLink parses the last page number out of the Link header,
// which looks like `<url>; rel="next", <url>; rel="last"`.
// It returns 0 if the header is absent or has no `last` link.
func parseLastPageFromLink(linkHeader string) int {
	matches := linkLastPageRegex.FindStringSubmatch(linkHeader)
	if len(matches) < 2 {
		return 0
	}

	lastPage, err := strconv.Atoi(matches[1])
	if err != nil {
		return 0
	}

	return lastPage
}
