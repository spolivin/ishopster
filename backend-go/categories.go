package main

import (
	"context"
	"errors"
	"log"
	"net/http"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Category struct {
	ID          int     `json:"id"`
	Name        string  `json:"name"`
	Slug        string  `json:"slug"`
	Description *string `json:"description"`
}

// listCategories returns every category, ordered by name.
func listCategories(ctx context.Context, pool *pgxpool.Pool) ([]Category, error) {
	rows, err := pool.Query(ctx, "SELECT id, name, slug, description FROM categories ORDER BY name")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	categories := []Category{}
	for rows.Next() {
		var c Category
		if err := rows.Scan(&c.ID, &c.Name, &c.Slug, &c.Description); err != nil {
			return nil, err
		}
		categories = append(categories, c)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return categories, nil
}

func getCategory(ctx context.Context, pool *pgxpool.Pool, slug string) (Category, error) {
	var c Category
	const q = `SELECT id, name, slug, description FROM categories WHERE slug = $1`
	err := pool.QueryRow(ctx, q, slug).Scan(&c.ID, &c.Name, &c.Slug, &c.Description)
	if err != nil {
		return Category{}, err
	}
	return c, nil
}

func categoriesHandler(pool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		categories, err := listCategories(r.Context(), pool)
		if err != nil {
			log.Printf("list categories: %v", err)
			writeJSON(w, http.StatusInternalServerError, errorResponse{
				Detail: "internal error",
			})
			return
		}
		writeJSON(w, http.StatusOK, categories)
	}
}

func categoryHandler(pool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		slug := r.PathValue("slug")
		category, err := getCategory(r.Context(), pool, slug)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				writeJSON(w, http.StatusNotFound, errorResponse{
					Detail: "category not found",
				})
				return
			}
			log.Printf("get category %q: %v", slug, err)
			writeJSON(w, http.StatusInternalServerError, errorResponse{
				Detail: "internal error",
			})
			return
		}
		writeJSON(w, http.StatusOK, category)
	}
}
