package main

import (
	"net/http"
	"os"
	"time"

	"github.com/caarlos0/starcharts/internal/github"

	"github.com/apex/httplog"
	"github.com/apex/log"
	"github.com/apex/log/handlers/text"
	"github.com/caarlos0/starcharts/config"
	"github.com/caarlos0/starcharts/controller"
	"github.com/caarlos0/starcharts/internal/cache"
	"github.com/go-redis/redis"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	log.SetHandler(text.New(os.Stderr))
	var config = config.Get()
	var ctx = log.WithField("port", config.Port)
	options, err := redis.ParseURL(config.RedisURL)
	if err != nil {
		log.WithError(err).Fatal("invalid redis_url")
	}
	var redis = redis.NewClient(options)
	var cache = cache.New(redis)
	defer cache.Close()
	var github = github.New(config, cache)

	var r = mux.NewRouter()
	r.Path("/").
		Methods(http.MethodGet).
		HandlerFunc(controller.Index())
	r.PathPrefix("/static/").
		Methods(http.MethodGet).
		Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	r.Path("/{owner}/{repo}.svg").
		Methods(http.MethodGet).
		HandlerFunc(controller.GetRepoChart(github, cache))
	r.Path("/{owner}/{repo}").
		Methods(http.MethodGet).
		HandlerFunc(controller.GetRepo(github, cache))

	// generic metrics
	var requestCounter = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "starcharts",
		Subsystem: "http",
		Name:      "requests_total",
		Help:      "total requests",
	}, []string{"code", "method"})
	var responseObserver = promauto.NewSummaryVec(prometheus.SummaryOpts{
		Namespace: "starcharts",
		Subsystem: "http",
		Name:      "responses",
		Help:      "response times and counts",
	}, []string{"code", "method"})
	prometheus.MustRegister(github.RateLimits)

	r.Methods(http.MethodGet).Path("/metrics").Handler(promhttp.Handler())

	var srv = &http.Server{
		Handler: httplog.New(
			promhttp.InstrumentHandlerDuration(
				responseObserver,
				promhttp.InstrumentHandlerCounter(
					requestCounter,
					r,
				),
			),
		),
		Addr:         "0.0.0.0:" + config.Port,
		WriteTimeout: 30 * time.Second,
		ReadTimeout:  30 * time.Second,
	}
	ctx.Info("starting up...")
	ctx.WithError(srv.ListenAndServe()).Error("failed to start up server")
}
