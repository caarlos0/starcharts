package github

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"regexp"
	"sort"
	"strconv"
	"sync"
	"time"

	"github.com/apex/log"
	"golang.org/x/sync/errgroup"
)

var (
	errNoMorePages = errors.New("no more pages to get")
	// linkLastPageRegex is used to parse the last page number from the Link header
	linkLastPageRegex = regexp.MustCompile(`[&?]page=(\d+)[^>]*>;\s*rel="last"`)
)

// maxConcurrentRequests is the maximum number of concurrent requests to GitHub API
const maxConcurrentRequests = 5

// Stargazer is a star at a given time.
type Stargazer struct {
	StarredAt time.Time `json:"starred_at"`
	// Count represents the actual position/count of this star (used in sampling mode).
	// If 0, use index+1 as count (non-sampling mode).
	Count int `json:"-"`
}

// Stargazers returns all the stargazers of a given repo.
// If star count is too large, it uses sampling mode to fetch data points.
func (gh *GitHub) Stargazers(ctx context.Context, repo Repository) (stars []Stargazer, err error) {
	// First request the first page to get the actual max page count (via Link header)
	firstPageStars, lastPage, err := gh.getFirstPageAndLastPage(ctx, repo)
	if err != nil {
		return nil, err
	}

	log.WithField("repo", repo.FullName).
		WithField("lastPage", lastPage).
		WithField("starCount", repo.StargazersCount).
		Debug("got pagination info from API")

	// If only one page or page count is less than max sample pages, fetch all pages
	if lastPage <= gh.maxSamplePages {
		return gh.getAllStargazersWithFirstPage(ctx, repo, firstPageStars, lastPage)
	}

	// Otherwise use sampling mode
	return gh.getSampledStargazers(ctx, repo, firstPageStars, lastPage)
}

// getFirstPageAndLastPage requests the first page and parses the Link header to get the max page count.
func (gh *GitHub) getFirstPageAndLastPage(ctx context.Context, repo Repository) ([]Stargazer, int, error) {
	log := log.WithField("repo", repo.FullName)

	resp, err := gh.makeStarPageRequest(ctx, repo, 1, "")
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusForbidden {
		rateLimits.Inc()
		log.Warn("rate limit hit")
		return nil, 0, ErrRateLimit
	}

	if resp.StatusCode != http.StatusOK {
		bts, _ := io.ReadAll(resp.Body)
		return nil, 0, fmt.Errorf("%w: %v", ErrGitHubAPI, string(bts))
	}

	bts, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, 0, err
	}

	var stars []Stargazer
	if err := json.Unmarshal(bts, &stars); err != nil {
		return nil, 0, err
	}

	// Parse Link header to get the max page count
	linkHeader := resp.Header.Get("Link")
	lastPage := gh.parseLastPageFromLink(linkHeader)

	// If no Link header or parsing failed, there is only one page
	if lastPage == 0 {
		lastPage = 1
	}

	log.WithField("lastPage", lastPage).Debug("parsed last page from Link header")

	return stars, lastPage, nil
}

// parseLastPageFromLink parses the max page count from the Link header.
// Link header format: <url>; rel="next", <url>; rel="last"
func (gh *GitHub) parseLastPageFromLink(linkHeader string) int {
	if linkHeader == "" {
		return 0
	}

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

// getAllStargazersWithFirstPage fetches all stargazers (used for small repositories).
// firstPageStars is the already fetched first page data.
func (gh *GitHub) getAllStargazersWithFirstPage(ctx context.Context, repo Repository, firstPageStars []Stargazer, lastPage int) (stars []Stargazer, err error) {
	stars = append(stars, firstPageStars...)

	// If only one page, return directly
	if lastPage <= 1 {
		return stars, nil
	}

	var (
		wg   errgroup.Group
		lock sync.Mutex
	)

	wg.SetLimit(maxConcurrentRequests)
	// Start fetching from page 2 (page 1 is already fetched)
	for page := 2; page <= lastPage; page++ {
		page := page
		wg.Go(func() error {
			result, err := gh.getStargazersPage(ctx, repo, page)
			if errors.Is(err, errNoMorePages) {
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
	err = wg.Wait()

	sort.Slice(stars, func(i, j int) bool {
		return stars[i].StarredAt.Before(stars[j].StarredAt)
	})
	return
}

// getSampledStargazers fetches stargazers using sampling mode (used for large repositories).
// Inspired by star-history project's sampling logic.
// firstPageStars is the already fetched first page data, lastPage is the actual max page count parsed from Link header.
func (gh *GitHub) getSampledStargazers(ctx context.Context, repo Repository, firstPageStars []Stargazer, lastPage int) (stars []Stargazer, err error) {
	log.WithField("repo", repo.FullName).
		WithField("lastPage", lastPage).
		Info("using sampling mode for large repo")

	// Calculate sample page numbers, evenly distributed across all pages
	samplePages := gh.calculateSamplePages(lastPage, gh.maxSamplePages)

	type pageResult struct {
		page      int
		star      Stargazer
		starCount int // the actual count position of this star
	}

	var (
		wg      errgroup.Group
		lock    sync.Mutex
		results []pageResult
	)

	// First page is already fetched, add it to results directly
	if len(firstPageStars) > 0 {
		results = append(results, pageResult{
			page:      1,
			star:      firstPageStars[0],
			starCount: 1,
		})
	}

	wg.SetLimit(maxConcurrentRequests)
	for _, page := range samplePages {
		// Skip first page (already fetched)
		if page == 1 {
			continue
		}
		page := page
		wg.Go(func() error {
			result, err := gh.getStargazersPage(ctx, repo, page)
			if errors.Is(err, errNoMorePages) {
				return nil
			}
			if err != nil {
				return err
			}
			if len(result) == 0 {
				return nil
			}

			// Calculate the actual position of the first star on this page (based on page number and page size)
			// The 1st star on page 1 is star #1
			// The 1st star on page N is star #(N-1)*pageSize + 1
			starCount := (page-1)*gh.pageSize + 1

			lock.Lock()
			defer lock.Unlock()
			results = append(results, pageResult{
				page:      page,
				star:      result[0],
				starCount: starCount,
			})
			return nil
		})
	}

	if err = wg.Wait(); err != nil {
		return nil, err
	}

	// Sort results by page number
	sort.Slice(results, func(i, j int) bool {
		return results[i].page < results[j].page
	})

	// Extract the first star from each sampled page as a data point and set Count
	for _, r := range results {
		star := r.star
		star.Count = r.starCount
		stars = append(stars, star)
	}

	// Add the last data point (current time and total star count)
	// This ensures the chart extends to the current time point
	stars = append(stars, Stargazer{
		StarredAt: time.Now(),
		Count:     repo.StargazersCount,
	})

	return stars, nil
}

// calculateSamplePages calculates the page numbers to sample.
// Evenly distributed across all pages, ensuring the first page is included.
func (gh *GitHub) calculateSamplePages(totalPages, maxSamples int) []int {
	pages := make([]int, 0, maxSamples)

	for i := 1; i <= maxSamples; i++ {
		// Calculate evenly distributed page numbers
		page := int(math.Round(float64(i*totalPages) / float64(maxSamples)))
		if page < 1 {
			page = 1
		}
		if page > totalPages {
			page = totalPages
		}
		pages = append(pages, page)
	}

	// Ensure first page is included (important for displaying start time)
	if len(pages) > 0 && pages[0] != 1 {
		pages[0] = 1
	}

	// Deduplicate (may have duplicates in edge cases)
	seen := make(map[int]bool)
	uniquePages := make([]int, 0, len(pages))
	for _, p := range pages {
		if !seen[p] {
			seen[p] = true
			uniquePages = append(uniquePages, p)
		}
	}

	return uniquePages
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
func (gh *GitHub) getStargazersPage(ctx context.Context, repo Repository, page int) ([]Stargazer, error) {
	log := log.WithField("repo", repo.FullName).WithField("page", page)
	defer log.Trace("get page").Stop(nil)

	var stars []Stargazer
	key := fmt.Sprintf("%s_%d", repo.FullName, page)
	etagKey := fmt.Sprintf("%s_%d", repo.FullName, page) + "_etag"

	var etag string
	if err := gh.cache.Get(etagKey, &etag); err != nil {
		log.WithError(err).Warnf("failed to get %s from cache", etagKey)
	}

	resp, err := gh.makeStarPageRequest(ctx, repo, page, etag)
	if err != nil {
		return stars, err
	}

	bts, err := io.ReadAll(resp.Body)
	if err != nil {
		return stars, err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusNotModified:
		effectiveEtags.Inc()
		log.Info("not modified")
		err := gh.cache.Get(key, &stars)
		if err != nil {
			log.WithError(err).Warnf("failed to get %s from cache", key)
			if err := gh.cache.Delete(etagKey); err != nil {
				log.WithError(err).Warnf("failed to delete %s from cache", etagKey)
			}
			return gh.getStargazersPage(ctx, repo, page)
		}
		return stars, err
	case http.StatusForbidden:
		rateLimits.Inc()
		log.Warn("rate limit hit")
		return stars, ErrRateLimit
	case http.StatusOK:
		if err := json.Unmarshal(bts, &stars); err != nil {
			return stars, err
		}
		if len(stars) == 0 {
			return stars, errNoMorePages
		}
		if err := gh.cache.Put(key, stars); err != nil {
			log.WithError(err).Warnf("failed to cache %s", key)
		}

		etag = resp.Header.Get("etag")
		if etag != "" {
			if err := gh.cache.Put(etagKey, etag); err != nil {
				log.WithError(err).Warnf("failed to cache %s", etagKey)
			}
		}

		return stars, nil
	default:
		return stars, fmt.Errorf("%w: %v", ErrGitHubAPI, string(bts))
	}
}

func (gh *GitHub) makeStarPageRequest(ctx context.Context, repo Repository, page int, etag string) (*http.Response, error) {
	url := fmt.Sprintf(
		"https://api.github.com/repos/%s/stargazers?page=%d&per_page=%d",
		repo.FullName,
		page,
		gh.pageSize,
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Accept", "application/vnd.github.v3.star+json")
	if etag != "" {
		req.Header.Add("If-None-Match", etag)
	}

	return gh.authorizedDo(req, 0)
}
