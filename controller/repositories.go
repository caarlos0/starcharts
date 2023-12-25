package controller

import (
	"fmt"
	"github.com/caarlos0/starcharts/internal/chart/svg"
	"io"
	"io/fs"
	"net/http"
	"time"

	"github.com/apex/log"
	"github.com/caarlos0/httperr"
	"github.com/caarlos0/starcharts/internal/cache"
	"github.com/caarlos0/starcharts/internal/chart"
	"github.com/caarlos0/starcharts/internal/github"
	"github.com/gorilla/mux"
)

const (
	CHART_WIDTH  = 1024
	CHART_HEIGHT = 400
)

// GetRepo shows the given repo chart.
func GetRepo(fsys fs.FS, github *github.GitHub, cache *cache.Redis, version string) http.Handler {
	return httperr.NewF(func(w http.ResponseWriter, r *http.Request) error {
		name := fmt.Sprintf(
			"%s/%s",
			mux.Vars(r)["owner"],
			mux.Vars(r)["repo"],
		)
		details, err := github.RepoDetails(r.Context(), name)
		if err != nil {
			return executeTemplate(fsys, w, map[string]error{
				"Error": err,
			})
		}
		return executeTemplate(fsys, w, map[string]interface{}{
			"Version": version,
			"Details": details,
		})
	})
}

var stylesMap = map[string]string{
	"light":    chart.LightStyles,
	"dark":     chart.DarkStyles,
	"adaptive": chart.AdaptiveStyles,
}

// GetRepoChart returns the SVG chart for the given repository.
//
// nolint: funlen
// TODO: refactor.
func GetRepoChart(gh *github.GitHub, cache *cache.Redis) http.Handler {
	return httperr.NewF(func(w http.ResponseWriter, r *http.Request) error {
		vars := mux.Vars(r)
		name := fmt.Sprintf("%s/%s", vars["owner"], vars["repo"])
		log := log.WithField("repo", name)
		defer log.Trace("collect_stars").Stop(nil)
		repo, err := gh.RepoDetails(r.Context(), name)
		if err != nil {
			return httperr.Wrap(err, http.StatusBadRequest)
		}

		params := r.URL.Query()

		stargazers, err := gh.Stargazers(r.Context(), repo)
		if err != nil {
			log.WithError(err).Error("failed to get stars")
			_, err = w.Write([]byte(errSvg(err)))
			return err
		}

		lineColor, err := extractColor(r, "line")
		if err != nil {
			return err
		}

		series := chart.Series{
			StrokeWidth: 2,
			Color:       lineColor,
		}
		for i, star := range stargazers {
			series.XValues = append(series.XValues, star.StarredAt)
			series.YValues = append(series.YValues, float64(i))
		}
		if len(series.XValues) < 2 {
			log.Info("not enough results, adding some fake ones")
			series.XValues = append(series.XValues, time.Now())
			series.YValues = append(series.YValues, 1)
		}

		backgroundColor, err := extractColor(r, "background")
		if err != nil {
			return err
		}

		axisColor, err := extractColor(r, "axis")
		if err != nil {
			return err
		}

		graph := chart.Chart{
			Width:      CHART_WIDTH,
			Height:     CHART_HEIGHT,
			Styles:     stylesMap[params.Get("variant")],
			Background: backgroundColor,
			XAxis: chart.XAxis{
				Name:        "Time",
				Color:       axisColor,
				StrokeWidth: 2,
			},
			YAxis: chart.YAxis{
				Name:        "Stargazers",
				Color:       axisColor,
				StrokeWidth: 2,
			},
			Series: series,
		}
		defer log.Trace("chart").Stop(&err)

		header := w.Header()
		header.Add("content-type", "image/svg+xml;charset=utf-8")
		header.Add("cache-control", "public, max-age=86400")
		header.Add("date", time.Now().Format(time.RFC1123))
		header.Add("expires", time.Now().Format(time.RFC1123))

		graph.Render(w)
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
