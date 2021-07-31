package controller

import (
	"html/template"
	"io/fs"
	"net/http"
)

func Index(fsys fs.FS) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := template.Must(template.ParseFS(fsys, "static/templates/index.html")).
			Execute(w, nil)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

func HandleForm(fsys fs.FS) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/"+r.FormValue("repository"), http.StatusSeeOther)
	}
}
