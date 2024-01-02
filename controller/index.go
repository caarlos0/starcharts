package controller

import (
	"html/template"
	"io/fs"
	"net/http"
	"strings"

	"github.com/caarlos0/httperr"
)

func Index(filesystem fs.FS, version string) http.Handler {
	indexTemplate, err := template.ParseFS(filesystem, base, index)
	if err != nil {
		panic(err)
	}

	return httperr.NewF(func(w http.ResponseWriter, r *http.Request) error {
		return indexTemplate.Execute(w, map[string]string{"Version": version})
	})
}

func HandleForm() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		repo := strings.TrimPrefix(r.FormValue("repository"), "https://github.com/")
		http.Redirect(w, r, repo, http.StatusSeeOther)
	}
}
