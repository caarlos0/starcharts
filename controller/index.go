package controller

import (
	"html/template"
	"io/fs"
	"net/http"
	"regexp"
	"strings"

	"github.com/caarlos0/httperr"
)

var repositoryExpression = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9-]*/[a-zA-Z0-9_.-]+$`)

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
		if !repositoryExpression.MatchString(repo) {
			http.Error(w, "invalid repository: expected owner/repository", http.StatusBadRequest)
			return
		}
		http.Redirect(w, r, "/"+repo, http.StatusSeeOther)
	}
}
