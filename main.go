package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"io/ioutil"

	"github.com/apex/httplog"
	"github.com/apex/log"
	"github.com/apex/log/handlers/text"
	cache "github.com/patrickmn/go-cache"
	chart "github.com/wcharczuk/go-chart"
)

const pageSize = 100

var (
	token       string
	port        string
	seriesCache *cache.Cache
)

type stargazer struct {
	StarredAt time.Time `json:"starred_at"`
}

func init() {
	log.SetHandler(text.New(os.Stderr))
	seriesCache = cache.New(1*time.Hour, 2*time.Hour)
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

func starchart(w http.ResponseWriter, r *http.Request) {
	var repo = r.URL.Path[1:]
	var ctx = log.WithField("repo", repo)
	if !strings.Contains(repo, "/") {
		http.Error(w, fmt.Sprintf("invalid repo: %s", repo), http.StatusBadRequest)
		return
	}
	series, err := collectStars(repo)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	seriesCache.Set(repo, series, cache.DefaultExpiration)
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
	cached, found := seriesCache.Get(name)
	if found {
		ctx.Info("got from cache")
		series = cached.(chart.TimeSeries)
		return
	}

	var page = 1
	for {
		ctx.Infof("getting page %d", page)
		url := fmt.Sprintf(
			"https://api.github.com/repos/%s/stargazers?page=%d&per_page=%d",
			name, page, pageSize,
		)
		req, err := http.NewRequest(http.MethodGet, url, nil)
		if err != nil {
			return series, err
		}
		req.Header.Add("Accept", "application/vnd.github.v3.star+json")
		if token != "" {
			req.Header.Add("Authorization", "token "+token)
		}
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return series, err
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			bts, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				return series, err
			}
			return series, fmt.Errorf("failed to get stargazers from github api: %v", string(bts))
		}
		var stargazers []stargazer
		if err := json.NewDecoder(resp.Body).Decode(&stargazers); err != nil {
			return series, err
		}
		if len(stargazers) == 0 {
			break
		}
		for i, star := range stargazers {
			series.XValues = append(series.XValues, star.StarredAt)
			series.YValues = append(series.YValues, float64(i+((page-1)*pageSize)))
		}
		page++
	}
	return
}
