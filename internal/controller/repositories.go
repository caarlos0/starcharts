package controller

import (
	"fmt"
	"html/template"
	"net/http"
	"time"

	"github.com/apex/log"
	"github.com/caarlos0/starchart/internal/cache"
	"github.com/caarlos0/starchart/internal/config"
	"github.com/caarlos0/starchart/internal/github"
	"github.com/gorilla/mux"
	chart "github.com/wcharczuk/go-chart"
)

// GetRepo shows the given repo chart
func GetRepo() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var owner = mux.Vars(r)["owner"]
		var repo = mux.Vars(r)["repo"]
		template.Must(template.ParseFiles("templates/index.html")).
			Execute(w, struct {
				Path string
			}{
				Path: fmt.Sprintf("%s/%s", owner, repo),
			})
	}
}

// GetRepoChart returns the SVG chart for the given repository
func GetRepoChart(cfg config.Config, cache *cache.Redis) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var name = fmt.Sprintf(
			"%s/%s",
			mux.Vars(r)["owner"],
			mux.Vars(r)["repo"],
		)
		var ctx = log.WithField("repo", name)
		defer ctx.Trace("collect_stars").Stop(nil)
		var github = github.New(cfg, cache)
		repo, err := github.RepoDetails(name)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		stargazers, err := github.Stargazers(repo)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		var series chart.TimeSeries
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
				Style:     chart.StyleShow(),
			},
			YAxis: chart.YAxis{
				Name:      "Sargazers",
				NameStyle: chart.StyleShow(),
				Style:     chart.StyleShow(),
			},
			Series: []chart.Series{series},
		}
		defer ctx.Trace("chart").Stop(&err)
		w.Header().Add("Content-Type", "image/svg+xml")
		graph.Render(chart.SVG, w)
	}
}
