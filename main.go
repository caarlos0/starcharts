package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	chart "github.com/wcharczuk/go-chart"
)

const pageSize = 100

var token string

type stargazer struct {
	StarredAt time.Time `json:"starred_at"`
}

type repository struct {
	FullName    string `json:"full_name"`
	Permissions struct {
		Push bool
	}
}

func main() {
	token = os.Getenv("GITHUB_TOKEN")
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		repo, err := getRepo(r.URL.Path[1:])
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if !repo.Permissions.Push {
			http.Error(w, "I do not have push permissions in this repo, won't spend my rate limit with it", http.StatusNotAcceptable)
			return
		}

		var series = chart.TimeSeries{}
		var page = 1
		for {
			url := fmt.Sprintf(
				"https://api.github.com/repos/%s/stargazers?page=%d&per_page=%d",
				repo.FullName, page, pageSize,
			)
			req, err := http.NewRequest(http.MethodGet, url, nil)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			req.Header.Add("Accept", "application/vnd.github.v3.star+json")
			if token != "" {
				req.Header.Add("Authorization", "token "+token)
			}
			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			defer resp.Body.Close()
			var stargazers []stargazer
			if err := json.NewDecoder(resp.Body).Decode(&stargazers); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
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
		graph := chart.Chart{
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
		w.Header().Add("Content-Type", "image/svg+xml")
		graph.Render(chart.SVG, w)
	})
	log.Fatalln(http.ListenAndServe(":3000", nil))
}

func getRepo(name string) (repo repository, err error) {
	if !strings.Contains(name, "/") {
		return repo, fmt.Errorf("invalid repo: %v", name)
	}
	url := fmt.Sprintf("https://api.github.com/repos/%s", name)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return
	}
	if token != "" {
		req.Header.Add("Authorization", "token "+token)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	err = json.NewDecoder(resp.Body).Decode(&repo)
	return
}
