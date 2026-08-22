package server

import (
	"io/fs"
	"net/http"
	"strings"
)

// assetHandler serves a static filesystem from an embedded fs.FS. It is a
// thin wrapper over http.FileServer (which exists in every Go 1.21) with
// explicit index.html resolution and content-type headers.
func assetHandler(sub fs.FS) http.Handler {
	base := http.FileServer(http.FS(sub))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/")
		// Directory request: only delegate to FileServer when the directory
		// actually contains an index.html (FileServer would otherwise show a
		// listing, which the console does not want).
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

// setAssetHeaders applies content-type and cache-control per asset class.
// The console is tiny and re-fetched often during development, so static
// assets are cached for a day while HTML and JSON are always revalidated.
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
