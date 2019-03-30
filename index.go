package main

import (
	"net/http"

	"github.com/caarlos0/starcharts/api"
)

func Handler(w http.ResponseWriter, r *http.Request) {
	api.Server().ServeHTTP(w, r)
}
