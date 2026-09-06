package main

import (
	"context"
	"log"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Address the HTTP server listens on.
const addr = ":8000"

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
	mux.HandleFunc("GET /api/categories", categoriesHandler(pool))
	mux.HandleFunc("GET /api/categories/{slug}", categoryHandler(pool))
	mux.HandleFunc("GET /api/products/{slug}", productHandler(pool))

	log.Printf("listening on %s", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}
