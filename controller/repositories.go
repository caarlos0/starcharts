package controller

import (
	"fmt"
	"html/template"
	"net/http"
	"time"

	"github.com/apex/log"
	"github.com/caarlos0/starcharts/internal/github"
	"github.com/gorilla/mux"
	chart "github.com/wcharczuk/go-chart"
	"github.com/wcharczuk/go-chart/drawing"
)

// GetRepo shows the given repo chart.
func GetRepo(github *github.GitHub) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		owner := mux.Vars(r)["owner"]
		name := mux.Vars(r)["repo"]
		details, err := github.RepoDetails(r.Context(), owner, name)
		if err != nil {
			log.WithError(err).Errorf("failed to get repo info")
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

// GetRepoChart returns the SVG chart for the given repository.
//
// nolint: funlen
// TODO: refactor.
func GetRepoChart(github *github.GitHub) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		owner := mux.Vars(r)["owner"]
		name := mux.Vars(r)["repo"]
		log := log.WithField("owner", owner).WithField("name", name)
		defer log.Trace("collect_stars").Stop(nil)
		stargazers, err := github.Stargazers(r.Context(), owner, name)
		if err != nil {
			log.WithError(err).Errorf("failed to plot chart")
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		series := chart.TimeSeries{
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
			log.Info("not enough results, adding some fake ones")
			series.XValues = append(series.XValues, time.Now())
			series.YValues = append(series.YValues, 1)
		}

		graph := chart.Chart{
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
		defer log.Trace("chart").Stop(&err)
		w.Header().Add("content-type", "image/svg+xml;charset=utf-8")
		w.Header().Add("cache-control", "public, max-age=86400")
		w.Header().Add("date", time.Now().Format(time.RFC1123))
		w.Header().Add("expires", time.Now().Format(time.RFC1123))
		if err := graph.Render(chart.SVG, w); err != nil {
			log.WithError(err).Error("failed to render graph")
		}
	}
}
