package main

import (
	"context"
	"errors"
	"log"
	"net/http"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Product struct {
	ID          int      `json:"id"`
	Name        string   `json:"name"`
	Slug        string   `json:"slug"`
	Description *string  `json:"description"`
	Price       string   `json:"price"`
	Category    Category `json:"category"`
}

func getProduct(ctx context.Context, pool *pgxpool.Pool, slug string) (Product, error) {
	var p Product
	const q = `
		SELECT p.id, p.name, p.slug, p.description, p.price,
       	       c.id, c.name, c.slug, c.description
		FROM products p
		JOIN categories c ON c.id = p.category_id
		WHERE p.slug = $1
		`
	err := pool.QueryRow(ctx, q, slug).Scan(
		&p.ID, &p.Name, &p.Slug, &p.Description, &p.Price,
		&p.Category.ID, &p.Category.Name, &p.Category.Slug, &p.Category.Description,
	)
	if err != nil {
		return Product{}, err
	}
	return p, nil
}

func productHandler(pool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		slug := r.PathValue("slug")
		product, err := getProduct(r.Context(), pool, slug)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				writeJSON(w, http.StatusNotFound, errorResponse{
					Detail: "product not found",
				})
				return
			}
			log.Printf("get product %q: %v", slug, err)
			writeJSON(w, http.StatusInternalServerError, errorResponse{
				Detail: "internal error",
			})
			return
		}
		writeJSON(w, http.StatusOK, product)
	}
}
