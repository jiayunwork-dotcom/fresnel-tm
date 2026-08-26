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

type Server struct {
	assets fs.FS
	mux    *http.ServeMux
}

func NewServer(assets fs.FS) *Server {
	s := &Server{assets: assets, mux: http.NewServeMux()}
	s.routes()
	return s
}

func (s *Server) routes() {
	s.mux.HandleFunc("/health", s.handleHealth)
	s.mux.HandleFunc("/api/health", s.handleHealth)
	s.mux.HandleFunc("/api/stack", s.handleStack)
	s.mux.HandleFunc("/api/spectrum", s.handleSpectrum)
	s.mux.HandleFunc("/api/examples", s.handleListExamples)
	s.mux.HandleFunc("/api/examples/", s.handleGetExample)

	if web, err := fs.Sub(s.assets, "web"); err == nil {
		s.mux.Handle("/", withMiddleware(assetHandler(web)))
	}
	if ex, err := fs.Sub(s.assets, "example"); err == nil {
		s.mux.Handle("/example/", withMiddleware(http.StripPrefix("/example/", assetHandler(ex))))
	}
}

func (s *Server) Handler() http.Handler {
	return withMiddleware(s.mux)
}

func (s *Server) ListenAndServe(addr string) error {
	srv := &http.Server{
		Addr:              addr,
		Handler:           s.Handler(),
		ReadHeaderTimeout: 10 * time.Second,
	}
	return srv.ListenAndServe()
}

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

func readBody(r *http.Request) ([]byte, error) {
	lr := http.MaxBytesReader(nil, r.Body, 1<<20)
	data, err := io.ReadAll(lr)
	if err != nil {
		return nil, fmt.Errorf("请求体过大或读取失败：%w", err)
	}
	return data, nil
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "service": "fresnel-tm"})
}

func jsonSnapshot(v any) string {
	b, err := json.Marshal(v)
	if err != nil {
		return fmt.Sprintf("<%v>", v)
	}
	return string(b)
}
