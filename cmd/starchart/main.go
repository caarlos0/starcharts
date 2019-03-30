package main

import (
	"net/http"
	"os"
	"time"

	"github.com/apex/log"
	"github.com/apex/log/handlers/text"
	"github.com/caarlos0/starcharts"
	"github.com/caarlos0/starcharts/config"
)

func init() {
	log.SetHandler(text.New(os.Stderr))
}

func main() {
	var config = config.Get()
	var ctx = log.WithField("port", config.Port)
	var srv = &http.Server{
		Handler:      starcharts.Server(),
		Addr:         "0.0.0.0:" + config.Port,
		WriteTimeout: 30 * time.Second,
		ReadTimeout:  30 * time.Second,
	}
	ctx.Info("starting up...")
	ctx.WithError(srv.ListenAndServe()).Error("failed to start up server")
}
