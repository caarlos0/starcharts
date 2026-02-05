package main

import (
	"embed"
	"net/http"
	"time"

	"github.com/caarlos0/starcharts/config"
	"github.com/caarlos0/starcharts/controller"
	"github.com/caarlos0/starcharts/internal/cache"
	"github.com/caarlos0/starcharts/internal/github"
	"github.com/charmbracelet/log"
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
	// log.SetLevel(log.DebugLevel)
	config := config.Get()
	ctx := log.With("listen", config.Listen)
	options, err := redis.ParseURL(config.RedisURL)
	if err != nil {
		log.Fatal("invalid redis_url", "err", err)
	}
	redis := redis.NewClient(options)
	cache := cache.New(redis)
	defer cache.Close()
	github := github.New(config, cache)

	r := mux.NewRouter()
	r.Path("/").
		Methods(http.MethodGet).
		Handler(controller.Index(static, version))
	r.Path("/").
		Methods(http.MethodPost).
		HandlerFunc(controller.HandleForm())
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
		Handler: promhttp.InstrumentHandlerDuration(
			responseObserver,
			promhttp.InstrumentHandlerCounter(
				requestCounter,
				r,
			),
		),
		Addr:         config.Listen,
		WriteTimeout: 60 * time.Second,
		ReadTimeout:  60 * time.Second,
	}
	ctx.Info("starting up...")
	if err := srv.ListenAndServe(); err != nil {
		ctx.Error("failed to start up server", "err", err)
	}
}
