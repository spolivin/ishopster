package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"strconv"

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

type ProductList struct {
	Items  []Product `json:"items"`
	Total  int       `json:"total"`
	Limit  int       `json:"limit"`
	Offset int       `json:"offset"`
}

func listProducts(ctx context.Context, pool *pgxpool.Pool, offset, limit int, category string) ([]Product, int, error) {
	const q = `
		SELECT p.id, p.name, p.slug, p.description, p.price,
       	       c.id, c.name, c.slug, c.description
		FROM products p
		JOIN categories c ON c.id = p.category_id
		WHERE ($1 = '' OR c.slug = $1)
		ORDER BY p.name, p.id
		LIMIT $2 OFFSET $3
		`
	rows, err := pool.Query(ctx, q, category, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	products := []Product{}
	for rows.Next() {
		var p Product
		err = rows.Scan(
			&p.ID, &p.Name, &p.Slug, &p.Description, &p.Price,
			&p.Category.ID, &p.Category.Name, &p.Category.Slug, &p.Category.Description,
		)
		if err != nil {
			return nil, 0, err
		}
		products = append(products, p)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}
	// Release the connection before the second query: holding two at once
	// lets concurrent requests exhaust the pool and wait on each other.
	rows.Close()

	var total int
	const qCount = `
		SELECT COUNT(*)
		FROM products p
		JOIN categories c ON c.id = p.category_id
		WHERE ($1 = '' OR c.slug = $1)
	`
	if err := pool.QueryRow(ctx, qCount, category).Scan(&total); err != nil {
		return nil, 0, err
	}

	return products, total, nil
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

func intParam(r *http.Request, name string, def, min, max int) (int, bool) {
	varStr := r.URL.Query().Get(name)
	if varStr == "" {
		return def, true
	}
	varInt, err := strconv.Atoi(varStr)
	if err != nil {
		return 0, false
	}
	if varInt < min || varInt > max {
		return 0, false
	}
	return varInt, true
}

func productsHandler(pool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		category := r.URL.Query().Get("category")
		limit, ok := intParam(r, "limit", 20, 1, 100)
		if !ok {
			writeJSON(w, http.StatusUnprocessableEntity, errorResponse{
				Detail: "limit must be an integer between 1 and 100",
			})
			return
		}
		offset, ok := intParam(r, "offset", 0, 0, 1_000_000)
		if !ok {
			writeJSON(w, http.StatusUnprocessableEntity, errorResponse{
				Detail: "offset must be a non-negative integer",
			})
			return
		}
		products, total, err := listProducts(r.Context(), pool, offset, limit, category)
		if err != nil {
			log.Printf("list products: %v", err)
			writeJSON(w, http.StatusInternalServerError, errorResponse{
				Detail: "internal error",
			})
			return
		}

		writeJSON(w, http.StatusOK, ProductList{
			Items:  products,
			Total:  total,
			Limit:  limit,
			Offset: offset,
		})
	}
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
