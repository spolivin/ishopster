-- SQL exercises for the ishopster catalog (categories, products).
-- Run against the dev database:
--   docker compose exec -T db psql -U dbuser -d dbname < sql/exercises.sql
-- or paste individual queries into `psql`.
--
-- Schema:
--   categories(id, name, slug UNIQUE, description, created_at, updated_at)
--   products(id, category_id -> categories.id ON DELETE RESTRICT,
--            name, slug UNIQUE, description, price NUMERIC(10,2),
--            created_at, updated_at)


-- ---------------------------------------------------------------------------
-- 2. Products of a single category, identified by the category slug.
--
-- INNER JOIN (not LEFT): products.category_id is NOT NULL with a foreign
-- key, so every product always has a matching category -- there is no
-- reason to keep rows with a missing category. The WHERE clause filters
-- to the one category we care about.
-- ---------------------------------------------------------------------------
SELECT p.name, c.slug
FROM products p
INNER JOIN categories c ON p.category_id = c.id
WHERE c.slug = 'laptops';


-- ---------------------------------------------------------------------------
-- 3. Number of products per category, including categories with zero.
--
-- LEFT JOIN from categories -> products: keep every category even when no
-- product matches (those rows get NULL product columns).
--
-- COUNT(p.id), not COUNT(*): for an empty category the LEFT JOIN still
-- produces one row with all product columns NULL. COUNT(*) counts that
-- row (-> 1); COUNT(p.id) counts non-NULL values of p.id (-> 0), which is
-- the real product count. COUNT(<column>) ignores NULLs; COUNT(*) does not.
-- ---------------------------------------------------------------------------
SELECT c.slug, COUNT(p.id) AS product_count
FROM categories c
LEFT JOIN products p ON c.id = p.category_id
GROUP BY c.slug
ORDER BY c.slug;


-- ---------------------------------------------------------------------------
-- 4. Products priced strictly between 1000 and 5000, cheapest first.
--
-- `>` / `<` exclude the bounds. BETWEEN would INCLUDE them
-- (price BETWEEN 1000 AND 5000 is 1000 <= price <= 5000).
-- ---------------------------------------------------------------------------
SELECT p.name, p.price, c.slug
FROM products p
INNER JOIN categories c ON p.category_id = c.id
WHERE p.price > 1000 AND p.price < 5000
ORDER BY p.price ASC;


-- ---------------------------------------------------------------------------
-- 5. Average product price per category, non-empty categories only.
--
-- INNER JOIN already drops categories with no products (no rows survive
-- the join, so there is nothing to group) -- no HAVING needed. HAVING
-- would only be required with a LEFT JOIN, to filter empty groups AFTER
-- aggregation (WHERE runs before GROUP BY and cannot do that).
--
-- ROUND(..., 2) trims AVG's long NUMERIC result to a currency scale.
-- ---------------------------------------------------------------------------
SELECT c.slug, ROUND(AVG(p.price), 2) AS avg_price
FROM categories c
INNER JOIN products p ON c.id = p.category_id
GROUP BY c.slug
ORDER BY c.slug;


-- ---------------------------------------------------------------------------
-- 6. Category holding the most expensive product.
--
-- Variant A -- subquery on MAX(price): returns EVERY product at the top
-- price, so a tie yields multiple rows. The subquery is uncorrelated
-- (runs once). `= (SELECT MAX(...))` is equivalent here since the value
-- is scalar.
SELECT c.slug, p.price
FROM products p
INNER JOIN categories c ON p.category_id = c.id
WHERE p.price = (SELECT MAX(price) FROM products);

-- Variant B -- ORDER BY ... DESC LIMIT 1: always exactly one row, chosen
-- arbitrarily among ties (non-deterministic without a tie-breaker).
SELECT c.slug, p.price
FROM products p
INNER JOIN categories c ON p.category_id = c.id
ORDER BY p.price DESC
LIMIT 1;

-- Variant C -- "top price, but keep all ties" (PostgreSQL 13+):
SELECT c.slug, p.price
FROM products p
INNER JOIN categories c ON p.category_id = c.id
ORDER BY p.price DESC
FETCH FIRST 1 ROW WITH TIES;
