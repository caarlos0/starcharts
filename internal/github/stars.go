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
	// 用于解析 Link header 中的 last 页码
	linkLastPageRegex = regexp.MustCompile(`[&?]page=(\d+)[^>]*>;\s*rel="last"`)
)

// maxConcurrentRequests 是并发请求 GitHub API 的最大数量
const maxConcurrentRequests = 5

// Stargazer is a star at a given time.
type Stargazer struct {
	StarredAt time.Time `json:"starred_at"`
	// Count 表示该 star 的实际位置/计数（用于采样模式）
	// 如果为 0，表示使用索引+1 作为计数（非采样模式）
	Count int `json:"-"`
}

// Stargazers returns all the stargazers of a given repo.
// 如果 star 数量太多，会采用采样方式获取数据点。
func (gh *GitHub) Stargazers(ctx context.Context, repo Repository) (stars []Stargazer, err error) {
	// 先请求第一页，获取实际可用的最大页数（通过 Link header）
	firstPageStars, lastPage, err := gh.getFirstPageAndLastPage(ctx, repo)
	if err != nil {
		return nil, err
	}

	log.WithField("repo", repo.FullName).
		WithField("lastPage", lastPage).
		WithField("starCount", repo.StargazersCount).
		Debug("got pagination info from API")

	// 如果只有一页或页数小于最大采样页数，获取所有页面
	if lastPage <= gh.maxSamplePages {
		return gh.getAllStargazersWithFirstPage(ctx, repo, firstPageStars, lastPage)
	}

	// 否则使用采样方式
	return gh.getSampledStargazers(ctx, repo, firstPageStars, lastPage)
}

// getFirstPageAndLastPage 请求第一页并解析 Link header 获取最大页数
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

	// 解析 Link header 获取最大页数
	linkHeader := resp.Header.Get("Link")
	lastPage := gh.parseLastPageFromLink(linkHeader)

	// 如果没有 Link header 或解析失败，说明只有一页
	if lastPage == 0 {
		lastPage = 1
	}

	log.WithField("lastPage", lastPage).Debug("parsed last page from Link header")

	return stars, lastPage, nil
}

// parseLastPageFromLink 从 Link header 解析出最大页数
// Link header 格式: <url>; rel="next", <url>; rel="last"
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

// getAllStargazersWithFirstPage 获取所有 stargazers（用于小型项目）
// firstPageStars 是已经获取的第一页数据
func (gh *GitHub) getAllStargazersWithFirstPage(ctx context.Context, repo Repository, firstPageStars []Stargazer, lastPage int) (stars []Stargazer, err error) {
	stars = append(stars, firstPageStars...)

	// 如果只有一页，直接返回
	if lastPage <= 1 {
		return stars, nil
	}

	var (
		wg   errgroup.Group
		lock sync.Mutex
	)

	wg.SetLimit(maxConcurrentRequests)
	// 从第 2 页开始获取（第 1 页已经有了）
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

// getSampledStargazers 使用采样方式获取 stargazers（用于大型项目）
// 参考 star-history 项目的采样逻辑
// firstPageStars 是已经获取的第一页数据，lastPage 是从 Link header 解析的实际最大页数
func (gh *GitHub) getSampledStargazers(ctx context.Context, repo Repository, firstPageStars []Stargazer, lastPage int) (stars []Stargazer, err error) {
	log.WithField("repo", repo.FullName).
		WithField("lastPage", lastPage).
		Info("using sampling mode for large repo")

	// 计算采样页码，均匀分布在所有页面中
	samplePages := gh.calculateSamplePages(lastPage, gh.maxSamplePages)

	type pageResult struct {
		page      int
		star      Stargazer
		starCount int // 该 star 的实际计数位置
	}

	var (
		wg      errgroup.Group
		lock    sync.Mutex
		results []pageResult
	)

	// 第一页已经有了，直接添加到结果中
	if len(firstPageStars) > 0 {
		results = append(results, pageResult{
			page:      1,
			star:      firstPageStars[0],
			starCount: 1,
		})
	}

	wg.SetLimit(maxConcurrentRequests)
	for _, page := range samplePages {
		// 跳过第一页（已经有了）
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

			// 计算该页第一个 star 的实际位置（基于页码和每页大小）
			// 第 1 页第 1 个 star 是第 1 颗星
			// 第 N 页第 1 个 star 是第 (N-1)*pageSize + 1 颗星
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

	// 按页码排序结果
	sort.Slice(results, func(i, j int) bool {
		return results[i].page < results[j].page
	})

	// 从每个采样页中提取第一个 star 作为数据点，并设置 Count
	for _, r := range results {
		star := r.star
		star.Count = r.starCount
		stars = append(stars, star)
	}

	// 添加最后一个数据点（当前时间和总 star 数）
	// 这样可以确保图表延伸到当前时间点
	stars = append(stars, Stargazer{
		StarredAt: time.Now(),
		Count:     repo.StargazersCount,
	})

	return stars, nil
}

// calculateSamplePages 计算需要采样的页码
// 均匀分布在所有页面中，确保包含第一页
func (gh *GitHub) calculateSamplePages(totalPages, maxSamples int) []int {
	pages := make([]int, 0, maxSamples)

	for i := 1; i <= maxSamples; i++ {
		// 计算均匀分布的页码
		page := int(math.Round(float64(i*totalPages) / float64(maxSamples)))
		if page < 1 {
			page = 1
		}
		if page > totalPages {
			page = totalPages
		}
		pages = append(pages, page)
	}

	// 确保第一页被包含（对于显示起始时间很重要）
	if len(pages) > 0 && pages[0] != 1 {
		pages[0] = 1
	}

	// 去重（可能在边界情况下有重复）
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
