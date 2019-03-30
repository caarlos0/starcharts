package starcharts

import (
	"net/http"
	"os"

	"github.com/apex/httplog"
	"github.com/apex/log"
	"github.com/apex/log/handlers/text"
	"github.com/caarlos0/starcharts/config"
	"github.com/caarlos0/starcharts/controller"
	"github.com/caarlos0/starcharts/internal/cache"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func init() {
	log.SetHandler(text.New(os.Stderr))
}

var singleton *http.Handler

func Server() http.Handler {
	if singleton != nil {
		return *singleton
	}
	log.Info("starting new server singleton")
	var config = config.Get()
	var cache = cache.New(config.RedisURL)
	defer cache.Close()

	var routes = mux.NewRouter()
	routes.Path("/").
		Methods(http.MethodGet).
		HandlerFunc(controller.Index())
	routes.PathPrefix("/static/").
		Methods(http.MethodGet).
		Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	routes.Path("/{owner}/{repo}.svg").
		Methods(http.MethodGet).
		HandlerFunc(controller.GetRepoChart(config, cache))
	routes.Path("/{owner}/{repo}").
		Methods(http.MethodGet).
		HandlerFunc(controller.GetRepo(config, cache))

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

	routes.Path("/metrics").
		Methods(http.MethodGet).
		Handler(promhttp.Handler())

	var handler http.Handler = httplog.New(
		promhttp.InstrumentHandlerDuration(
			responseObserver,
			promhttp.InstrumentHandlerCounter(
				requestCounter,
				routes,
			),
		),
	)
	singleton = &handler
	return handler
}
