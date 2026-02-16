package webServer

import (
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/WoodPKrs/GoogleCalendarTimetable/database"
)

// EndpointContext holds shared dependencies for HTTP handlers that are not
// part of the ConnectRPC service (e.g. the static frontend server).
type EndpointContext struct {
	Query *database.Queries
}

var fs = http.FileServer(http.Dir("./static/"))

func (ec *EndpointContext) ServeFrontend(w http.ResponseWriter, r *http.Request) {
	// Prevent caching of env.js since it's dynamically generated at runtime
	if r.URL.Path == "/env.js" {
		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
		w.Header().Set("Pragma", "no-cache")
		w.Header().Set("Expires", "0")
	}

	if r.URL.Path != "/" {
		fullPath := "./static/" + strings.TrimPrefix(path.Clean(r.URL.Path), "/")
		_, err := os.Stat(fullPath)
		if err != nil {
			if !os.IsNotExist(err) {
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				return
			}
			// Requested file does not exist so we return the default (resolves to index.html)
			r.URL.Path = "/"
		}
	}
	fs.ServeHTTP(w, r)
}
