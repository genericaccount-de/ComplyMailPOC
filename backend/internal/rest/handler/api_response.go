package handler

import (
	"encoding/json"
	"log"
	"net/http"
)

// maxRequestBytes bounds the size of incoming JSON request bodies to guard
// against oversized payloads.
const maxRequestBytes = 1 << 20 // 1 MiB

// errorResponse is the JSON shape returned for error conditions.
type errorResponse struct {
	Error string `json:"error"`
}

// writeJSON serializes v as JSON with the given status code.
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("handler: encode response: %v", err)
	}
}

// writeError writes a JSON error body with the given status code.
func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, errorResponse{Error: msg})
}
