package main

import (
	"encoding/json"
	"log"
	"net/http"
)

// healthResponse is the body of a successful health check.
type healthResponse struct {
	Status string `json:"status"`
}

// errorResponse is the body of any error reply. Same shape as FastAPI's
// HTTPException detail, so both backends stay interchangeable.
type errorResponse struct {
	Detail string `json:"detail"`
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	err := json.NewEncoder(w).Encode(v)
	if err != nil {
		log.Printf("writeJSON: %v", err)
	}
}
