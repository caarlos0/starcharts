package controller

import (
	"html/template"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/apex/log"
)

func Index() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		dir, err := os.Getwd()
		if err != nil {
			log.WithError(err).Error("failed to read dir")
		}
		log.WithField("cwd", dir).Info("current dir")
		files, err := ioutil.ReadDir(dir)
		if err != nil {
			log.WithError(err).Error("failed to read dir")
		}
		for _, f := range files {
			log.Info(f.Name())
		}
		template.Must(template.ParseFiles("templates/index.html")).Execute(w, nil)
	}
}
