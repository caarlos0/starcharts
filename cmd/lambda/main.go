package main

import (
	"net/http"

	"github.com/caarlos0/starcharts"
)

func Handler(w http.ResponseWriter, r *http.Request) {
	starcharts.Server().ServeHTTP(w, r)
}
