package controller

import (
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
	"regexp"
	"time"
)

const (
	base       = "static/templates/base.gohtml"
	repository = "static/templates/repository.gohtml"
	index      = "static/templates/index.gohtml"
)

var colorExpression = regexp.MustCompile("^#([a-fA-F0-9]{6}|[a-fA-F0-9]{3}|[a-fA-F0-9]{8})$")

func extractColor(r *http.Request, name string) (string, error) {
	color := r.URL.Query().Get(name)
	if len(color) == 0 {
		return "", nil
	}

	if colorExpression.MatchString(color) {
		return color, nil
	}

	return "", fmt.Errorf("invalid %s: %s", name, color)
}

type params struct {
	Owner      string
	Repo       string
	Line       string
	Background string
	Axis       string
	Variant    string
}

func extractSvgChartParams(r *http.Request) (*params, error) {
	backgroundColor, err := extractColor(r, "background")
	if err != nil {
		return nil, err
	}

	axisColor, err := extractColor(r, "axis")
	if err != nil {
		return nil, err
	}

	lineColor, err := extractColor(r, "line")
	if err != nil {
		return nil, err
	}

	vars := mux.Vars(r)

	return &params{
		Owner:      vars["owner"],
		Repo:       vars["repo"],
		Background: backgroundColor,
		Axis:       axisColor,
		Line:       lineColor,
		Variant:    r.URL.Query().Get("variant"),
	}, nil
}

func writeSvgHeaders(w http.ResponseWriter) {
	header := w.Header()
	header.Add("content-type", "image/svg+xml;charset=utf-8")
	header.Add("cache-control", "public, max-age=86400")
	header.Add("date", time.Now().Format(time.RFC1123))
	header.Add("expires", time.Now().Format(time.RFC1123))
}

func chartKey(params *params) string {
	return fmt.Sprintf(
		"%s/%s/[%s][%s][%s][%s]",
		params.Owner,
		params.Repo,
		params.Variant,
		params.Background,
		params.Axis,
		params.Line,
	)
}
