package github

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/apex/log"
)

// Repository details.
type Repository struct {
	FullName        string `json:"full_name"`
	StargazersCount int    `json:"stargazers_count"`
	CreatedAt       string `json:"created_at"`
}

var ErrorNotFound = errors.New("Repository not found")

// RepoDetails gets the given repository details.
// 获取给定的存储库详细信息。
func (gh *GitHub) RepoDetails(ctx context.Context, name string) (Repository, error) {
	var repo Repository
	log := log.WithField("repo", name)

	var etag string
	etagKey := name + "_etag" // 标记

	// 先找缓存
	if err := gh.cache.Get(etagKey, &etag); err != nil {
		log.WithError(err).Warnf("failed to get %s from cache", etagKey)
	}

	// http请求，去github获取响应，内部分为未经授权的请求和携带令牌的请求
	resp, err := gh.makeRepoRequest(ctx, name, etag)
	if err != nil {
		return repo, err
	}

	bts, err := io.ReadAll(resp.Body)
	if err != nil {
		return repo, err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusNotModified: // 304
		log.Info("not modified") // 未修改
		effectiveEtags.Inc()     // 有效标记
		err := gh.cache.Get(name, &repo)
		if err != nil {
			log.WithError(err).Warnf("failed to get %s from cache", name)
			if err := gh.cache.Delete(etagKey); err != nil {
				log.WithError(err).Warnf("failed to delete %s from cache", etagKey)
			}
			return gh.RepoDetails(ctx, name)
		}
		return repo, err
	case http.StatusForbidden: // 403
		rateLimits.Inc()
		log.Warn("rate limit hit")
		return repo, ErrRateLimit
	case http.StatusOK: // 200
		// 响应结果正确，解析成功，存入缓存
		if err := json.Unmarshal(bts, &repo); err != nil {
			return repo, err
		}
		if err := gh.cache.Put(name, repo); err != nil {
			log.WithError(err).Warnf("failed to cache %s", name)
		}

		etag = resp.Header.Get("etag") // 获取标记，用标记作为Key再存一次缓存
		if etag != "" {
			if err := gh.cache.Put(etagKey, etag); err != nil {
				log.WithError(err).Warnf("failed to cache %s", etagKey)
			}
		}

		return repo, nil
	case http.StatusNotFound: // 404 没有找到
		return repo, ErrorNotFound
	default: // github api 有问题
		return repo, fmt.Errorf("%w: %v", ErrGitHubAPI, string(bts))
	}
}

func (gh *GitHub) makeRepoRequest(ctx context.Context, name, etag string) (*http.Response, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s", name)
	// http请求获取svg图片
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	if etag != "" {
		req.Header.Add("If-None-Match", etag)
	}

	return gh.authorizedDo(req, 0) // 获取github令牌，并http请求
}
