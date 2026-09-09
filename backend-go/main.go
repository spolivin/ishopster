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

// statusRecorder wraps a ResponseWriter to remember the status code sent to
// the client. The interface can be written to but not read back, so the code
// has to be captured on its way out.
//
// The embedded ResponseWriter supplies Header and Write unchanged, which is
// what makes the wrapper usable anywhere the real one is.
type statusRecorder struct {
	http.ResponseWriter
	status int
}

// WriteHeader records the status code, then lets the real ResponseWriter send
// it. Delegating through the embedded field is required: calling
// rec.WriteHeader here would recurse into this method forever.
func (rec *statusRecorder) WriteHeader(status int) {
	rec.status = status
	rec.ResponseWriter.WriteHeader(status)
}

// logging wraps a handler so that every request is logged with its method,
// URI, status and duration. It wraps the whole mux rather than individual
// handlers, so responses the router generates itself (404, 405) are logged
// too.
func logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rec, r)
		elapsed := time.Since(start)
		log.Printf("%s %s %d %s", r.Method, r.URL.RequestURI(), rec.status, elapsed)
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
