package webServer

import (
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"

	"github.com/woodpkrs/PlanerPK/config"
)

var (
	staticFileServer http.Handler
	fileServerOnce   sync.Once
)

func getFileServer() http.Handler {
	fileServerOnce.Do(func() {
		staticFileServer = http.FileServer(http.Dir(config.Config.StaticDir))
	})
	return staticFileServer
}

func ServeFrontend(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		fullPath := filepath.Join(config.Config.StaticDir, strings.TrimPrefix(path.Clean(r.URL.Path), "/"))
		fi, err := os.Stat(fullPath)
		if err != nil {
			if !os.IsNotExist(err) {
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				return
			}
			r.URL.Path = "/"
		} else if fi.IsDir() {
			r.URL.Path = "/"
		}
	}
	getFileServer().ServeHTTP(w, r)
}
