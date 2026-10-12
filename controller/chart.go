package controller

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

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
func GetRepoChart(gh *github.GitHub, cache *cache.Redis) http.Handler {
	return httperr.NewF(func(w http.ResponseWriter, r *http.Request) error {
		p, err := extractSvgChartParams(r)
		if err != nil {
			slog.Error("failed to extract params", "error", err)
			return err
		}

		cacheKey := chartKey(p)
		name := fmt.Sprintf("%s/%s", p.Owner, p.Repo)
		log := slog.With("repo", name, "variant", p.Variant)

		if served, err := serveFromCache(w, cache, cacheKey, log); served {
			return err
		}

		return renderChart(r, w, gh, cache, p, name, cacheKey, log)
	})
}

// serveFromCache writes a chart straight from cache if one exists.
// The bool return says whether the request was already handled.
func serveFromCache(w http.ResponseWriter, cache *cache.Redis, cacheKey string, log *slog.Logger) (bool, error) {
	cachedChart := ""
	if err := cache.Get(cacheKey, &cachedChart); err != nil {
		return false, nil
	}
	writeSvgHeaders(w)
	log.Debug("using cached chart")
	_, err := fmt.Fprint(w, cachedChart)
	return true, err
}

// renderChart fetches stargazers, builds the SVG chart, writes it to w,
// and caches the result.
func renderChart(
	r *http.Request,
	w http.ResponseWriter,
	gh *github.GitHub,
	cache *cache.Redis,
	p *params,
	name, cacheKey string,
	log *slog.Logger,
) error {
	start := time.Now()
	repo, err := gh.RepoDetails(r.Context(), name)
	if err != nil {
		return httperr.Wrap(err, http.StatusBadRequest)
	}

	stargazers, err := gh.Stargazers(r.Context(), repo)
	if err != nil {
		log.Error("failed to get stars", "error", err)
		writeSvgHeaders(w)
		_, err = w.Write([]byte(errSvg(err)))
		return err
	}
	log.Debug("collect_stars", "duration", time.Since(start))

	chartStart := time.Now()
	graph := &chart.Chart{
		Width:      CHART_WIDTH,
		Height:     CHART_HEIGHT,
		Styles:     stylesMap[p.Variant],
		Background: p.Background,
		XAxis: chart.XAxis{
			Name:        "Time",
			Color:       p.Axis,
			StrokeWidth: 2,
		},
		YAxis: chart.YAxis{
			Name:        "Stargazers",
			Color:       p.Axis,
			StrokeWidth: 2,
		},
		Series: buildSeries(stargazers, p, log),
	}

	writeSvgHeaders(w)
	cacheBuffer := &strings.Builder{}
	graph.Render(io.MultiWriter(w, cacheBuffer))
	log.Debug("chart", "duration", time.Since(chartStart))

	if err := cache.Put(cacheKey, cacheBuffer.String()); err != nil {
		log.Error("failed to cache chart", "error", err)
	}
	return nil
}

// buildSeries converts stargazers into chart series data, falling back to
// two synthetic points when there isn't enough data to draw a line.
func buildSeries(stargazers []github.Stargazer, p *params, log *slog.Logger) chart.Series {
	series := chart.Series{
		StrokeWidth: 2,
		Color:       p.Line,
	}

	for i, star := range stargazers {
		series.XValues = append(series.XValues, star.StarredAt)
		// If star.Count > 0, use the actual count from sampling mode.
		// Otherwise use index+1 (non-sampling mode, continuous data).
		if star.Count > 0 {
			series.YValues = append(series.YValues, float64(star.Count))
		} else {
			series.YValues = append(series.YValues, float64(i+1))
		}
	}

	if len(series.XValues) < 2 {
		log.Info("not enough results, adding some fake ones")
		series.XValues = append(series.XValues, time.Now())
		series.YValues = append(series.YValues, 1)
	}

	return series
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
