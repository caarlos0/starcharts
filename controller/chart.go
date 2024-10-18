package controller

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/apex/log"
	"github.com/caarlos0/httperr"
	"github.com/caarlos0/starcharts/internal/cache"
	"github.com/caarlos0/starcharts/internal/chart"
	"github.com/caarlos0/starcharts/internal/chart/svg"
	"github.com/caarlos0/starcharts/internal/github"
)

var stylesMap = map[string]string{
	"light":    chart.LightStyles,
	"dark":     chart.DarkStyles,
	"adaptive": chart.AdaptiveStyles,
}

// GetRepoChart returns the SVG chart for the given repository.
// 返回给定存储库的SVG图表。
// nolint: funlen
// TODO: refactor.
func GetRepoChart(gh *github.GitHub, cache *cache.Redis) http.Handler {
	return httperr.NewF(func(w http.ResponseWriter, r *http.Request) error {
		params, err := extractSvgChartParams(r) // 提取svg图片的请求参数
		if err != nil {
			log.WithError(err).Error("failed to extract params")
			return err
		}

		cacheKey := chartKey(params) // 拼接参数得到缓存使用的key值
		name := fmt.Sprintf("%s/%s", params.Owner, params.Repo)
		log := log.WithField("repo", name).WithField("variant", params.Variant)

		cachedChart := "" // 查找缓存
		if err = cache.Get(cacheKey, &cachedChart); err == nil {
			// 使用缓存中的svg图片
			writeSvgHeaders(w) // 拼接svg图片响应的响应头信息
			log.Debug("using cached chart")
			_, err := fmt.Fprint(w, cachedChart)
			return err
		}

		// Trace返回一个带有Stop方法的新条目，以触发相应的完成日志，这对延迟很有用。
		defer log.Trace("collect_stars").Stop(nil)

		// 缓存中没有对应svg图片

		// 去github请求获取数据
		repo, err := gh.RepoDetails(r.Context(), name)
		if err != nil {
			return httperr.Wrap(err, http.StatusBadRequest)
		}
		// 从响应中获取star情况
		stargazers, err := gh.Stargazers(r.Context(), repo)
		if err != nil {
			log.WithError(err).Error("failed to get stars")
			writeSvgHeaders(w)
			_, err = w.Write([]byte(errSvg(err)))
			return err
		}

		series := chart.Series{
			StrokeWidth: 2,
			Color:       params.Line,
		}
		for i, star := range stargazers {
			series.XValues = append(series.XValues, star.StarredAt)
			series.YValues = append(series.YValues, float64(i+1))
		}
		if len(series.XValues) < 2 {
			log.Info("not enough results, adding some fake ones") // 没有足够的结果，添加一些假的
			series.XValues = append(series.XValues, time.Now())
			series.YValues = append(series.YValues, 1)
		}

		graph := &chart.Chart{ // 图标数据拼接
			Width:      CHART_WIDTH,
			Height:     CHART_HEIGHT,
			Styles:     stylesMap[params.Variant],
			Background: params.Background,
			XAxis: chart.XAxis{
				Name:        "Time",
				Color:       params.Axis,
				StrokeWidth: 2,
			},
			YAxis: chart.YAxis{
				Name:        "Stargazers",
				Color:       params.Axis,
				StrokeWidth: 2,
			},
			Series: series,
		}
		defer log.Trace("chart").Stop(&err)

		writeSvgHeaders(w) // 拼接svg图片响应的响应头信息

		cacheBuffer := &strings.Builder{}
		// io.MultiWriter可以将多个io.Writer组合成一个，这样写入数据时会同时写入到所有组合的Writer中。
		// 在这里，同时向w和cacheBuffer进行写入。
		graph.Render(io.MultiWriter(w, cacheBuffer))
		err = cache.Put(cacheKey, cacheBuffer.String()) // 存入缓存
		if err != nil {
			log.WithError(err).Error("failed to cache chart")
		}

		return nil
	})
}

func errSvg(err error) string {
	return svg.SVG().
		Attr("width", svg.Px(CHART_WIDTH)).
		Attr("height", svg.Px(CHART_HEIGHT)).
		ContentFunc(func(writer io.Writer) {
			svg.Text().
				Attr("fill", "red").
				Attr("x", svg.Px(CHART_WIDTH/2)).
				Attr("y", svg.Px(CHART_HEIGHT/2)).
				Content(err.Error()).
				Render(writer)
		}).
		String()
}
