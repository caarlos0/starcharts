package main

import (
	"net/http"
	"os"
	"time"

	"github.com/apex/httplog"
	"github.com/apex/log"
	"github.com/apex/log/handlers/text"
	"github.com/caarlos0/starcharts/config"
	"github.com/caarlos0/starcharts/controller"
	"github.com/caarlos0/starcharts/internal/cache"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func init() {
	log.SetHandler(text.New(os.Stderr))
}

func main() {
	var config = config.Get()
	var ctx = log.WithField("port", config.Port)
	var cache = cache.New(config.RedisURL)
	defer cache.Close()

	var r = mux.NewRouter()
	r.Path("/").
		Methods(http.MethodGet).
		HandlerFunc(controller.Index())
	r.Path("/metrics").
		Methods(http.MethodGet).
		Handler(promhttp.Handler())
	r.PathPrefix("/static/").
		Methods(http.MethodGet).
		Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	r.Path("/{owner}/{repo}.svg").
		Methods(http.MethodGet).
		HandlerFunc(controller.GetRepoChart(config, cache))
	r.Path("/{owner}/{repo}").
		Methods(http.MethodGet).
		HandlerFunc(controller.GetRepo(config, cache))

	var srv = &http.Server{
		Handler:      httplog.New(r),
		Addr:         "0.0.0.0:" + config.Port,
		WriteTimeout: 30 * time.Second,
		ReadTimeout:  30 * time.Second,
	}
	ctx.Info("starting up...")
	ctx.WithError(srv.ListenAndServe()).Error("failed to start up server")
}
