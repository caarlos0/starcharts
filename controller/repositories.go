package controller

import (
	"fmt"
	"html/template"
	"net/http"
	"time"

	"github.com/apex/log"
	"github.com/caarlos0/starcharts/internal/cache"
	"github.com/caarlos0/starcharts/internal/github"
	"github.com/gorilla/mux"
	chart "github.com/wcharczuk/go-chart"
	"github.com/wcharczuk/go-chart/drawing"
)

// GetRepo shows the given repo chart
func GetRepo(github *github.GitHub, cache *cache.Redis) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var name = fmt.Sprintf(
			"%s/%s",
			mux.Vars(r)["owner"],
			mux.Vars(r)["repo"],
		)
		details, err := github.RepoDetails(name)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		err = template.Must(template.ParseFiles("templates/index.html")).
			Execute(w, details)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

// IntValueFormatter is a ValueFormatter for int.
func IntValueFormatter(v interface{}) string {
	return fmt.Sprintf("%.0f", v)
}

// GetRepoChart returns the SVG chart for the given repository
// nolint: funlen
// TODO: refactor
func GetRepoChart(github *github.GitHub, cache *cache.Redis) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var name = fmt.Sprintf(
			"%s/%s",
			mux.Vars(r)["owner"],
			mux.Vars(r)["repo"],
		)
		var ctx = log.WithField("repo", name)
		defer ctx.Trace("collect_stars").Stop(nil)
		repo, err := github.RepoDetails(name)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		stargazers, err := github.Stargazers(repo)
		if err != nil {
			http.Error(w, err.Error(), http.StatusServiceUnavailable)
			return
		}
		var series = chart.TimeSeries{
			Style: chart.Style{
				Show: true,
				StrokeColor: drawing.Color{
					R: 129,
					G: 199,
					B: 239,
					A: 255,
				},
				StrokeWidth: 2,
			},
		}
		for i, star := range stargazers {
			series.XValues = append(series.XValues, star.StarredAt)
			series.YValues = append(series.YValues, float64(i))
		}
		if len(series.XValues) < 2 {
			ctx.Info("not enough results, adding some fake ones")
			series.XValues = append(series.XValues, time.Now())
			series.YValues = append(series.YValues, 1)
		}

		var graph = chart.Chart{
			XAxis: chart.XAxis{
				Name:      "Time",
				NameStyle: chart.StyleShow(),
				Style: chart.Style{
					Show:        true,
					StrokeWidth: 2,
					StrokeColor: drawing.Color{
						R: 85,
						G: 85,
						B: 85,
						A: 255,
					},
				},
			},
			YAxis: chart.YAxis{
				Name:      "Stargazers",
				NameStyle: chart.StyleShow(),
				Style: chart.Style{
					Show:        true,
					StrokeWidth: 2,
					StrokeColor: drawing.Color{
						R: 85,
						G: 85,
						B: 85,
						A: 255,
					},
				},
				ValueFormatter: IntValueFormatter,
			},
			Series: []chart.Series{series},
		}
		defer ctx.Trace("chart").Stop(&err)
		w.Header().Add("content-type", "image/svg+xml;charset=utf-8")
		w.Header().Add("cache-control", "public, max-age=86400")
		w.Header().Add("date", time.Now().Format(time.RFC1123))
		w.Header().Add("expires", time.Now().Format(time.RFC1123))
		if err := graph.Render(chart.SVG, w); err != nil {
			log.WithError(err).Error("failed to render graph")
		}
	}
}
