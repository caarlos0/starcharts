package controller

import (
	"embed"
	"html/template"
	"net/http"
)

func Index(fs embed.FS) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var err = template.Must(template.ParseFS(fs, "static/templates/index.html")).
			Execute(w, nil)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}
