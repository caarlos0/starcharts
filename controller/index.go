package controller

import (
	"html/template"
	"io"
	"io/fs"
	"net/http"

	"github.com/caarlos0/httperr"
)

func Index(fsys fs.FS) http.Handler {
	return httperr.NewF(func(w http.ResponseWriter, r *http.Request) error {
		return executeTemplate(fsys, w, nil)
	})
}

func HandleForm(fsys fs.FS) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/"+r.FormValue("repository"), http.StatusSeeOther)
	}
}

func executeTemplate(fsys fs.FS, w io.Writer, data interface{}) error {
	return template.Must(template.ParseFS(fsys, "static/templates/index.html")).
		Execute(w, data)
}
