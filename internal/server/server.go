// Package server exposes the fresnel-tm solver over HTTP: the two JSON
// endpoints POST /api/stack and POST /api/spectrum plus a static web console
// served from an embedded filesystem.
//
// Every handler follows the same contract:
//
//   - valid requests return 200 with the solver output;
//   - malformed JSON, unknown fields and physically impossible inputs return
//     400 with {"error": "..."} describing the exact violation;
//   - solver failures (which the validation rules should prevent) return 422;
//   - panics inside handlers are recovered into 500 error bodies.
//
// The error body is produced by the backend and rendered verbatim by the
// frontend, so a bad request is always visible on the page, never silently
// swallowed.
package server

import (
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"log"
	"net/http"
	"time"
)

// Server wires the API handlers and the embedded web/example assets.
type Server struct {
	assets fs.FS // embedded filesystem containing web/ and example/
	mux    *http.ServeMux
}

// NewServer builds a server over the given embedded filesystem. The FS must
// contain a "web" directory (the console) and an "example" directory (the
// packaged example files).
func NewServer(assets fs.FS) *Server {
	s := &Server{assets: assets, mux: http.NewServeMux()}
	s.routes()
	return s
}

// routes registers every endpoint on the server's mux. The mux uses plain
// paths (no method patterns, no path wildcards) to stay compatible with
// Go 1.21; each handler checks the HTTP method itself.
func (s *Server) routes() {
	s.mux.HandleFunc("/api/stack", s.handleStack)
	s.mux.HandleFunc("/api/spectrum", s.handleSpectrum)
	s.mux.HandleFunc("/api/examples", s.handleListExamples)
	s.mux.HandleFunc("/api/examples/", s.handleGetExample)

	// Static assets: / serves the web console, /example/ serves the packaged
	// examples so the page can fetch them with plain fetch().
	if web, err := fs.Sub(s.assets, "web"); err == nil {
		s.mux.Handle("/", withMiddleware(assetHandler(web)))
	}
	if ex, err := fs.Sub(s.assets, "example"); err == nil {
		s.mux.Handle("/example/", withMiddleware(http.StripPrefix("/example/", assetHandler(ex))))
	}
}

// Handler returns the fully wired http.Handler.
func (s *Server) Handler() http.Handler {
	return withMiddleware(s.mux)
}

// ListenAndServe starts the server and blocks until it fails. The returned
// error is non-nil when the listener cannot be bound or the server crashes.
func (s *Server) ListenAndServe(addr string) error {
	srv := &http.Server{
		Addr:              addr,
		Handler:           s.Handler(),
		ReadHeaderTimeout: 10 * time.Second,
	}
	return srv.ListenAndServe()
}

// withMiddleware wraps a handler with request logging and panic recovery so
// that a handler bug still surfaces as a JSON error instead of a dropped
// connection.
func withMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		defer func() {
			if rec := recover(); rec != nil {
				log.Printf("panic in %s %s: %v", r.Method, r.URL.Path, rec)
				writeError(w, http.StatusInternalServerError, "服务器内部错误")
			}
		}()
		next.ServeHTTP(w, r)
		log.Printf("%s %s (%s)", r.Method, r.URL.Path, time.Since(start).Round(time.Millisecond))
	})
}

// readBody reads a JSON request body with a size cap. A body larger than
// 1 MiB is rejected so the handler stays cheap.
func readBody(r *http.Request) ([]byte, error) {
	lr := http.MaxBytesReader(nil, r.Body, 1<<20)
	data, err := io.ReadAll(lr)
	if err != nil {
		return nil, fmt.Errorf("请求体过大或读取失败：%w", err)
	}
	return data, nil
}

// jsonSnapshot formats a value for request logging.
func jsonSnapshot(v any) string {
	b, err := json.Marshal(v)
	if err != nil {
		return fmt.Sprintf("<%v>", v)
	}
	return string(b)
}
