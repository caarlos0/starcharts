package controller

import (
	"html/template"
	"io"
	"io/fs"
	"net/http"
	"strings"

	"github.com/caarlos0/httperr"
)

func Index(fsys fs.FS, version string) http.Handler {
	return httperr.NewF(func(w http.ResponseWriter, r *http.Request) error {
		return executeTemplate(fsys, w, map[string]string{"Version": version})
	})
}

func HandleForm(fsys fs.FS) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		repo := strings.TrimPrefix(r.FormValue("repository"), "https://github.com/")
		http.Redirect(w, r, repo, http.StatusSeeOther)
	}
}

func executeTemplate(fsys fs.FS, w io.Writer, data interface{}) error {
	return template.Must(template.ParseFS(fsys, "static/templates/index.html")).
		Execute(w, data)
}
