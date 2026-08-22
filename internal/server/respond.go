package server

import (
	"encoding/json"
	"net/http"
)

// errorBody is the wire format of every failure response. The message field
// is written by the validation layer and is meant to be shown to the user
// verbatim, e.g. "层厚不能为负，实际为 -10 nm".
type errorBody struct {
	Error string `json:"error"`
}

// writeError sends a JSON error body with the given status code. The Content-
// Type is set before writing so browsers render the body as JSON.
func writeError(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(errorBody{Error: msg})
}

// writeJSON sends a successful JSON response with status 200 (or the given
// code) and the canonical content type.
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// methodNotAllowed writes a 405 for routes that exist but reject the method.
// It is wired as a safety net for future endpoints.
func methodNotAllowed(w http.ResponseWriter) {
	writeError(w, http.StatusMethodNotAllowed, "该路径不支持此请求方法")
}

// notFound writes a JSON 404 so API consumers always get a structured error.
func notFound(w http.ResponseWriter) {
	writeError(w, http.StatusNotFound, "未找到该资源")
}
