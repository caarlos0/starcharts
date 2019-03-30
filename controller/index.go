package controller

import "net/http"
import "html/template"

func Index() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		template.Must(template.New("index").Parse(index)).Execute(w, nil)
	}
}
