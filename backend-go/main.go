package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Address the HTTP server listens on.
const addr = ":8000"

// Middleware logging every request
func logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		elapsed := time.Since(start)
		log.Printf("%s %s %s", r.Method, r.URL.RequestURI(), elapsed)
	})
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
	mux.HandleFunc("GET /api/categories", categoriesHandler(pool))
	mux.HandleFunc("GET /api/categories/{slug}", categoryHandler(pool))
	mux.HandleFunc("GET /api/products", productsHandler(pool))
	mux.HandleFunc("GET /api/products/{slug}", productHandler(pool))

	log.Printf("listening on %s", addr)
	log.Fatal(http.ListenAndServe(addr, logging(mux)))
}
