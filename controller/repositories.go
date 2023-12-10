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

// IntValueFormatter is a ValueFormatter for int.
func IntValueFormatter(v interface{}) string {
	return fmt.Sprintf("%.0f", v)
}

// GetRepoChart returns the SVG chart for the given repository.
//
// nolint: funlen
// TODO: refactor.
func GetRepoChart(gh *github.GitHub, cache *cache.Redis) http.Handler {
	return httperr.NewF(func(w http.ResponseWriter, r *http.Request) error {
		name := fmt.Sprintf(
			"%s/%s",
			mux.Vars(r)["owner"],
			mux.Vars(r)["repo"],
		)
		log := log.WithField("repo", name)
		defer log.Trace("collect_stars").Stop(nil)
		repo, err := gh.RepoDetails(r.Context(), name)
		if err != nil {
			return httperr.Wrap(err, http.StatusBadRequest)
		}

		w.Header().Add("content-type", "image/svg+xml;charset=utf-8")
		w.Header().Add("cache-control", "public, max-age=86400")
		w.Header().Add("date", time.Now().Format(time.RFC1123))
		w.Header().Add("expires", time.Now().Format(time.RFC1123))

		stargazers, err := gh.Stargazers(r.Context(), repo)
		if err != nil {
			log.WithError(err).Error("failed to get stars")
			_, err = w.Write([]byte(errSvg(err)))
			return err
		}
		series := chart.Series{
			StrokeWidth: 2,
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

		graph := chart.Chart{
			Width:  1024,
			Height: 400,
			XAxis: chart.XAxis{
				Name:        "Time",
				StrokeWidth: 2,
			},
			YAxis: chart.YAxis{
				Name:        "Stargazers",
				StrokeWidth: 2,
			},
			Series: series,
		}
		defer log.Trace("chart").Stop(&err)
		graph.Render(w)
		return nil
	})
}

func errSvg(err error) string {
	return svg.SVG().
		Attr("width", svg.Px(1024)).
		Attr("height", svg.Px(50)).
		ContentFunc(func(writer io.Writer) {
			svg.Text().
				Attr("fill", "red").
				Attr("x", svg.Px(100)).
				Attr("y", svg.Px(20)).
				Content(err.Error()).
				Render(writer)
		}).
		String()
}
