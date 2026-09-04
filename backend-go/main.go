package main

import (
	"encoding/json"
	"log"
	"net/http"
)

// Address the HTTP server listens on.
const addr = ":8000"

type Response struct {
	Status string `json:"status"`
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	err := json.NewEncoder(w).Encode(v)
	if err != nil {
		log.Printf("writeJSON: %v", err)
	}
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	healthResponse := Response{
		Status: "ok",
	}
	writeJSON(w, http.StatusOK, healthResponse)
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", healthHandler)
	mux.HandleFunc("GET /health/db", healthHandler)

	config := loadConfig()

	log.Printf("Host: %s, Port: %s, DB: %s", config.Host, config.Port, config.DB)
	log.Printf("listening on %s", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}
