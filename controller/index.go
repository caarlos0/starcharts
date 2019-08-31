package controller

import "net/http"
import "html/template"

func Index() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var err = template.Must(template.ParseFiles("templates/index.html")).
			Execute(w, nil)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}
