package main

import (
	"log"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
)

// healthHandler reports that the process itself is alive (liveness).
func healthHandler(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, healthResponse{Status: "ok"})
}

// healthDBHandler reports whether the database is reachable (readiness).
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
