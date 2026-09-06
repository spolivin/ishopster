package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Address the HTTP server listens on.
const addr = ":8000"

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

func healthHandler(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, healthResponse{Status: "ok"})
}

func healthDBHandler(pool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := pool.Ping(r.Context()); err != nil {
			log.Printf("could not ping database: %v", err)
			writeJSON(w, http.StatusServiceUnavailable, errorResponse{
				Detail: "database unreachable",
			})
			return
		}
		writeJSON(w, http.StatusOK, healthResponse{Status: "ok"})
	}
}

func main() {
	config := loadConfig()
	log.Printf("Host: %s, Port: %s, DB: %s", config.Host, config.Port, config.DB)

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, config.DSN())
	if err != nil {
		log.Fatalf("unable to create connection pool: %v", err)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		log.Fatalf("could not ping database: %v", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", healthHandler)
	mux.HandleFunc("GET /health/db", healthDBHandler(pool))

	log.Printf("listening on %s", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}
