package main

import (
	"embed"
	"math/rand"
	"net/http"
	"os"
	"time"

	"github.com/apex/httplog"
	"github.com/apex/log"
	"github.com/apex/log/handlers/text"
	"github.com/caarlos0/starcharts/config"
	"github.com/caarlos0/starcharts/controller"
	"github.com/caarlos0/starcharts/internal/cache"
	"github.com/caarlos0/starcharts/internal/chart"
	"github.com/caarlos0/starcharts/internal/github"
	"github.com/go-redis/redis"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

//go:embed static/*
var static embed.FS

var version = "devel"

func main() {
	log.SetHandler(text.New(os.Stderr))
	// log.SetLevel(log.DebugLevel)
	config := config.Get()
	ctx := log.WithField("listen", config.Listen)
	options, err := redis.ParseURL(config.RedisURL)
	if err != nil {
		log.WithError(err).Fatal("invalid redis_url")
	}
	redis := redis.NewClient(options)
	cache := cache.New(redis)
	defer cache.Close()
	github := github.New(config, cache)

	r := mux.NewRouter()
	r.Path("/").
		Methods(http.MethodGet).
		Handler(controller.Index(static, version))
	r.Path("/demo").
		Methods(http.MethodGet).
		Handler(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {

			series := chart.Series{}
			for i := 0; i < 100; i++ {
				series.XValues = append(series.XValues, time.Now().Add(time.Hour*24*time.Duration(i)))
				series.YValues = append(series.YValues, float64(i)+float64(i/4)*rand.Float64())
			}

			c := chart.Chart{
				Width:  1024,
				Height: 400,
				XAxis:  chart.XAxis{Name: "Time"},
				YAxis:  chart.YAxis{Name: "Stargazers"},
				Series: series,
			}

			writer.Header().Set("Content-Type", "image/svg+xml")
			c.Render(writer)
		}))
	r.Path("/").
		Methods(http.MethodPost).
		HandlerFunc(controller.HandleForm(static))
	r.PathPrefix("/static/").
		Methods(http.MethodGet).
		Handler(http.FileServer(http.FS(static)))
	r.Path("/{owner}/{repo}.svg").
		Methods(http.MethodGet).
		Handler(controller.GetRepoChart(github, cache))
	r.Path("/{owner}/{repo}").
		Methods(http.MethodGet).
		Handler(controller.GetRepo(static, github, cache, version))

	// generic metrics
	requestCounter := promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "starcharts",
		Subsystem: "http",
		Name:      "requests_total",
		Help:      "total requests",
	}, []string{"code", "method"})
	responseObserver := promauto.NewSummaryVec(prometheus.SummaryOpts{
		Namespace: "starcharts",
		Subsystem: "http",
		Name:      "responses",
		Help:      "response times and counts",
	}, []string{"code", "method"})

	r.Methods(http.MethodGet).Path("/metrics").Handler(promhttp.Handler())

	srv := &http.Server{
		Handler: httplog.New(
			promhttp.InstrumentHandlerDuration(
				responseObserver,
				promhttp.InstrumentHandlerCounter(
					requestCounter,
					r,
				),
			),
		),
		Addr:         config.Listen,
		WriteTimeout: 60 * time.Second,
		ReadTimeout:  60 * time.Second,
	}
	ctx.Info("starting up...")
	ctx.WithError(srv.ListenAndServe()).Error("failed to start up server")
}
