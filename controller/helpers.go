package controller

import (
	"fmt"
	"net/http"
	"regexp"
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
