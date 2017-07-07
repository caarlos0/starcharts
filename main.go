package main

import (
	"html/template"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/apex/httplog"
	"github.com/apex/log"
	"github.com/apex/log/handlers/text"
	"github.com/caarlos0/starchart/internal/github"
	chart "github.com/wcharczuk/go-chart"
)

const pageSize = 100

var (
	token string
	port  string
)

type stargazer struct {
	StarredAt time.Time `json:"starred_at"`
}

func init() {
	log.SetHandler(text.New(os.Stderr))
	token = os.Getenv("GITHUB_TOKEN")
	port = os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}
}

func main() {
	var mux = http.NewServeMux()
	mux.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		// ignored
	})
	mux.HandleFunc("/", starchart)
	var ctx = log.WithField("port", port)
	ctx.Info("starting up")
	if err := http.ListenAndServe(":"+port, httplog.New(mux)); err != nil {
		ctx.Fatal("failed to start up")
	}
}

var index = `<html>
<head>
	<title>StarChart</title>
</head>
<body>
	<p>
		Not a valid repository full name.
	</p>
	<p>
		Try <a href="goreleaser/goreleaser">/goreleaser/goreleaser</a>,
		for example
	</p>
</body>
</html>`

var tmpl = template.Must(template.New("index").Parse(index))

func starchart(w http.ResponseWriter, r *http.Request) {
	var repo = r.URL.Path[1:]
	var ctx = log.WithField("repo", repo)
	if !strings.Contains(repo, "/") {
		w.WriteHeader(http.StatusNotFound)
		if err := tmpl.Execute(w, nil); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	series, err := collectStars(repo)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
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

func collectStars(name string) (series chart.TimeSeries, err error) {
	var ctx = log.WithField("repo", name)
	defer ctx.Trace("collect_stars").Stop(&err)
	repo, err := github.RepoDetails(token, name)
	if err != nil {
		return
	}
	stargazers, err := github.Stargazers(token, repo)
	if err != nil {
		return
	}
	for i, star := range stargazers {
		series.XValues = append(series.XValues, star.StarredAt)
		series.YValues = append(series.YValues, float64(i))
	}
	return
}
