package controller

import (
	"fmt"
	"github.com/caarlos0/starcharts/internal/chart/svg"
	"html/template"
	"io"
	"io/fs"
	"net/http"
	"strings"
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
	repositoryTemplate, err := template.ParseFS(fsys, "static/templates/base.gohtml", "static/templates/repository.gohtml")
	if err != nil {
		panic(err)
	}

	errorTemplate, err := template.ParseFS(fsys, "static/templates/base.gohtml", "static/templates/error.gohtml")
	if err != nil {
		panic(err)
	}

	return httperr.NewF(func(w http.ResponseWriter, r *http.Request) error {
		name := fmt.Sprintf(
			"%s/%s",
			mux.Vars(r)["owner"],
			mux.Vars(r)["repo"],
		)
		details, err := github.RepoDetails(r.Context(), name)
		if err != nil {
			return errorTemplate.Execute(w, map[string]error{
				"Error": err,
			})
		}

		return repositoryTemplate.Execute(w, map[string]interface{}{
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
		params, err := extractParams(r)
		if err != nil {
			return err
		}
		cacheKey := chartKey(params)

		name := fmt.Sprintf("%s/%s", params.Owner, params.Repo)
		log := log.WithField("repo", name)

		cachedChart := ""
		if err = cache.Get(cacheKey, cachedChart); err == nil {
			_, err := fmt.Fprint(w, cachedChart)

			return err
		}

		defer log.Trace("collect_stars").Stop(nil)
		repo, err := gh.RepoDetails(r.Context(), name)
		if err != nil {
			return httperr.Wrap(err, http.StatusBadRequest)
		}

		stargazers, err := gh.Stargazers(r.Context(), repo)
		if err != nil {
			log.WithError(err).Error("failed to get stars")
			_, err = w.Write([]byte(errSvg(err)))
			return err
		}

		series := chart.Series{
			StrokeWidth: 2,
			Color:       params.Line,
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

		graph := &chart.Chart{
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

		header := w.Header()
		header.Add("content-type", "image/svg+xml;charset=utf-8")
		header.Add("cache-control", "public, max-age=86400")
		header.Add("date", time.Now().Format(time.RFC1123))
		header.Add("expires", time.Now().Format(time.RFC1123))

		cacheBuffer := strings.Builder{}
		graph.Render(io.MultiWriter(w, &cacheBuffer))
		err = cache.Put(cacheKey, cacheBuffer.String())
		if err != nil {
			log.WithError(err).Error("failed to cache chart")
		}

		return nil
	})
}

type Params struct {
	Owner      string
	Repo       string
	Line       string
	Background string
	Axis       string
	Variant    string
}

func extractParams(r *http.Request) (*Params, error) {
	backgroundColor, err := extractColor(r, "background")
	if err != nil {
		return nil, err
	}

	axisColor, err := extractColor(r, "axis")
	if err != nil {
		return nil, err
	}

	lineColor, err := extractColor(r, "line")
	if err != nil {
		return nil, err
	}

	vars := mux.Vars(r)

	return &Params{
		Owner:      vars["owner"],
		Repo:       vars["repo"],
		Background: backgroundColor,
		Axis:       axisColor,
		Line:       lineColor,
		Variant:    r.URL.Query().Get("variant"),
	}, nil
}

func chartKey(params *Params) string {
	return fmt.Sprintf(
		"%s_%s_%s_%s_%s_%s",
		params.Owner,
		params.Repo,
		params.Variant,
		params.Line,
		params.Background,
		params.Axis,
	)
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
