package controller

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestHandleForm(t *testing.T) {
	for _, tt := range []struct {
		name       string
		repository string
		status     int
		location   string
	}{
		{name: "repository", repository: "owner/repo", status: http.StatusSeeOther, location: "/owner/repo"},
		{name: "github URL", repository: "https://github.com/owner/repo", status: http.StatusSeeOther, location: "/owner/repo"},
		{name: "absolute URL", repository: "https://redirect.invalid/path", status: http.StatusBadRequest},
		{name: "network path", repository: "//redirect.invalid/path", status: http.StatusBadRequest},
		{name: "prefixed network path", repository: "https://github.com///redirect.invalid/path", status: http.StatusBadRequest},
		{name: "backslash", repository: `owner\repo/name`, status: http.StatusBadRequest},
		{name: "traversal", repository: "../repo", status: http.StatusBadRequest},
		{name: "query", repository: "owner/repo?next=1", status: http.StatusBadRequest},
		{name: "empty", status: http.StatusBadRequest},
	} {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(url.Values{"repository": {tt.repository}}.Encode()))
			r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			w := httptest.NewRecorder()
			HandleForm().ServeHTTP(w, r)
			if w.Code != tt.status || w.Header().Get("Location") != tt.location {
				t.Fatalf("got status %d, Location %q; want %d, %q", w.Code, w.Header().Get("Location"), tt.status, tt.location)
			}
		})
	}
}
