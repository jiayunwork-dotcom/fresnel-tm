package server

import (
	"io/fs"
	"net/http"
	"strings"
)

func assetHandler(sub fs.FS) http.Handler {
	base := http.FileServer(http.FS(sub))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/")
		index := "index.html"
		if path != "" {
			index = strings.TrimSuffix(path, "/") + "/index.html"
		}
		if path == "" || strings.HasSuffix(path, "/") {
			if _, err := fs.Stat(sub, index); err != nil {
				http.NotFound(w, r)
				return
			}
		}
		setAssetHeaders(w, path)
		base.ServeHTTP(w, r)
	})
}

func setAssetHeaders(w http.ResponseWriter, path string) {
	switch {
	case strings.HasSuffix(path, ".html"):
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("Cache-Control", "no-cache")
	case strings.HasSuffix(path, ".js"):
		w.Header().Set("Content-Type", "text/javascript; charset=utf-8")
		w.Header().Set("Cache-Control", "public, max-age=86400")
	case strings.HasSuffix(path, ".css"):
		w.Header().Set("Content-Type", "text/css; charset=utf-8")
		w.Header().Set("Cache-Control", "public, max-age=86400")
	case strings.HasSuffix(path, ".json"):
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.Header().Set("Cache-Control", "no-cache")
	default:
		w.Header().Set("Content-Type", "application/octet-stream")
	}
}
