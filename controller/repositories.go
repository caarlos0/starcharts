package controller

import (
	"fmt"
	"github.com/caarlos0/httperr"
	"github.com/caarlos0/starcharts/internal/cache"
	"github.com/caarlos0/starcharts/internal/github"
	"github.com/gorilla/mux"
	"html/template"
	"io/fs"
	"net/http"
)

const (
	CHART_WIDTH  = 1024
	CHART_HEIGHT = 400
)

// GetRepo shows the given repo chart.
func GetRepo(fsys fs.FS, gh *github.GitHub, cache *cache.Redis, version string) http.Handler {
	repositoryTemplate, err := template.ParseFS(fsys, "static/templates/base.gohtml", "static/templates/repository.gohtml")
	if err != nil {
		panic(err)
	}

	intexTemplate, err := template.ParseFS(fsys, "static/templates/base.gohtml", "static/templates/index.gohtml")
	if err != nil {
		panic(err)
	}

	return httperr.NewF(func(w http.ResponseWriter, r *http.Request) error {
		name := fmt.Sprintf(
			"%s/%s",
			mux.Vars(r)["owner"],
			mux.Vars(r)["repo"],
		)
		details, err := gh.RepoDetails(r.Context(), name)
		if err != nil {
			return intexTemplate.Execute(w, map[string]error{
				"Error": err,
			})
		}

		return repositoryTemplate.Execute(w, map[string]interface{}{
			"Version": version,
			"Details": details,
		})
	})
}
