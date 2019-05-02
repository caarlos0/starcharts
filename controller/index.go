package controller

import (
	"html/template"
	"net/http"
)

func Index() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		template.Must(template.ParseFiles("templates/index.html")).Execute(w, nil)
	}
}
