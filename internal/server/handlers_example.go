package server

import (
	"io/fs"
	"net/http"
	"sort"
	"strings"
)

// ExampleMeta is the lightweight listing entry returned by /api/examples.
// It deliberately omits the full layer data; the console fetches a single
// example with GET /api/examples/{name} when the user picks one.
type ExampleMeta struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	DesignNm    float64 `json:"design_wavelength_nm,omitempty"`
}

// handleListExamples serves GET /api/examples: the names of every packaged
// example file, sorted alphabetically, with their metadata.
func (s *Server) handleListExamples(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w)
		return
	}
	names, err := s.exampleNames()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	metas := make([]ExampleMeta, 0, len(names))
	for _, name := range names {
		metas = append(metas, ExampleMeta{Name: name})
	}
	writeJSON(w, http.StatusOK, map[string]any{"examples": metas})
}

// handleGetExample serves GET /api/examples/{name}: the full example body as
// stored on disk. The console posts this body straight to /api/stack.
func (s *Server) handleGetExample(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w)
		return
	}
	name := exampleNameFromPath(r.URL.Path)
	if !validExampleName(name) {
		writeError(w, http.StatusBadRequest, "示例名只能包含小写字母、数字、连字符与下划线")
		return
	}
	data, err := s.readExample(name)
	if err != nil {
		writeError(w, http.StatusNotFound, "没有名为 "+name+" 的示例")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
}

// exampleNameFromPath extracts the example name from a /api/examples/… path.
func exampleNameFromPath(path string) string {
	const prefix = "/api/examples/"
	if len(path) <= len(prefix) {
		return ""
	}
	return path[len(prefix):]
}

// exampleNames lists the JSON example files under example/.
func (s *Server) exampleNames() ([]string, error) {
	dir, err := fs.ReadDir(s.assets, "example")
	if err != nil {
		return nil, err
	}
	var names []string
	for _, e := range dir {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") {
			continue
		}
		names = append(names, strings.TrimSuffix(e.Name(), ".json"))
	}
	sort.Strings(names)
	return names, nil
}

// readExample returns the raw bytes of example/{name}.json.
func (s *Server) readExample(name string) ([]byte, error) {
	return fs.ReadFile(s.assets, "example/"+name+".json")
}

// validExampleName keeps example names to the safe character set so a crafted
// path can never escape the embedded example directory.
func validExampleName(name string) bool {
	if name == "" {
		return false
	}
	for _, r := range name {
		switch {
		case r >= 'a' && r <= 'z':
		case r >= '0' && r <= '9':
		case r == '-' || r == '_':
		default:
			return false
		}
	}
	return true
}
