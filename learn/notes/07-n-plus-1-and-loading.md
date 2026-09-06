# Lesson 7 — N+1 and loading strategies

Both backends answer `GET /api/products/{slug}` with the product **and** its
category nested inside. Getting related data out of a relational database is
the single most common place where a correct-looking service becomes slow.

This note is about that: what N+1 is, why an ORM invites it, and what changes
when you write the SQL yourself.

---

## 1. The problem

You want a list of products, each with its category name. Two ways to get it:

```
One query:                          N+1 queries:
                                    SELECT * FROM products            → N rows
SELECT p.*, c.*                     SELECT * FROM categories WHERE id=1
FROM products p                     SELECT * FROM categories WHERE id=2
JOIN categories c                   SELECT * FROM categories WHERE id=1   ← again!
  ON c.id = p.category_id           … once per product
```

The second shape is called **N+1**: one query for the list, plus one per row.
Note the repetition — the same category is fetched over and over, because
nothing remembers that it was already loaded.

**The cost is per round-trip, not per row.** Same data, same result; the only
difference is how many times the process talks to the database. Measured
locally against a small table, a 1-query version versus an 11-query version
comes out roughly 10× apart — the ratio tracks the query count, not the data
volume. Over a network link each extra round-trip adds its own latency, so the
gap widens with distance.

> When you measure this yourself, run a warm-up call first and average over
> many iterations. The very first query of a process pays for connection setup
> and statement preparation, which is easily larger than the effect you are
> trying to measure — it can even make the *faster* version look slower.

---

## 2. Why an ORM invites N+1

In SQLAlchemy, a relationship is not a field — it is a descriptor with
behaviour:

```python
products = (await session.execute(select(Product))).scalars().all()  # 1 query
for p in products:
    print(p.category.name)   # ← each access can issue its own SELECT
```

There is no line here that *looks* like a database call. `p.category` reads
like an attribute access, but if the relationship was not loaded, the ORM goes
and fetches it. That is **lazy loading**, and it is the default.

This is what makes N+1 dangerous rather than merely bad: it is invisible at the
call site, it does not fail, and it only shows up as latency under real data.

### Eager loading: telling the ORM not to be lazy

| Strategy         | Queries | Shape                                                 |
| ---------------- | ------- | ----------------------------------------------------- |
| lazy (default)   | 1 + N   | one query per accessed relationship                    |
| `selectinload`   | 2       | second query: `… WHERE id IN (…)`, results distributed |
| `joinedload`     | 1       | a JOIN; parent columns repeat per child row            |

`backend/app/routers/products.py` uses `selectinload(Product.category)`, which
turns 1 + N into 2. It is a *cure for laziness* — a concept that only exists
because the laziness exists.

---

## 3. What changes without an ORM

In `backend-go/products.go` the nested category is a plain struct field:

```go
type Product struct {
    …
    Category Category `json:"category"`   // data, no behaviour
}
```

It cannot query anything. It holds whatever `Scan` put there, and nothing else.
Reading `p.Category.Name` in a loop touches memory, never the database.

So **N+1 cannot happen by accident**. To create one you have to write it:

```go
for _, p := range products {
    var c Category
    pool.QueryRow(ctx, "SELECT … WHERE id=$1", p.CategoryID).Scan(&c…)  // ← visible
}
```

A query inside a loop is something a reader sees immediately. The failure mode
moves from *invisible* to *obvious*.

The price is that nobody chooses the loading strategy for you.

---

## 4. Choosing the strategy by hand

| | JOIN (one query) | Two queries (`selectinload` by hand) |
| --- | --- | --- |
| Round-trips | 1 | 2 |
| Duplicated data | parent columns repeat per child row | none |
| Code | one `Scan` per row | second query + a `map` to distribute results |
| Best for | many-to-one (product → category) | one-to-many, or a wide parent row |

For **many-to-one** — every product has exactly one category — a JOIN is the
obvious answer. The row count does not change; only a few extra columns ride
along. That is what both backends effectively do for a single product.

For **one-to-many** the JOIN starts multiplying rows. A product with three
images arrives as three rows, and you must fold them back together in code,
tracking which parent you have already seen. That is precisely the case where
the two-query approach reads better:

```
SELECT … FROM products WHERE …                    → collect ids
SELECT … FROM images WHERE product_id IN (…)      → group into map[int][]Image
```

Which is exactly what `selectinload` does internally — you are just writing it
out.

---

## 5. Spotting it

- **SQLAlchemy:** create the engine with `echo=True` (or enable the
  `sqlalchemy.engine` logger) and count the statements for one request. A page
  that logs one `SELECT` per item in a list is the signature.
- **Any backend:** `pg_stat_statements` in PostgreSQL, or simply counting
  queries in a test — assert that one request issues one query.
- **In review:** look for a database call inside a loop, and for attribute
  access on ORM relationships that were never eagerly loaded.

---

## Reading exercise

1. Why is `p.category.name` in a Python loop dangerous, while `p.Category.Name`
   in a Go loop is not?
2. `selectinload` makes 2 queries and `joinedload` makes 1. Why is the 2-query
   version usually the default recommendation for collections?
3. A JOIN fetches everything in one round-trip. Why not always use a JOIN?
4. Your service and the database run on the same machine during development,
   and on separate hosts in production. Why does an N+1 that "seemed fine"
   locally become a problem after deploying?
5. You add product images (one product → many images) and keep a single JOIN.
   What breaks in the result set, and what must the scanning loop now do?

---

## Answers

**1.** Because in Python the relationship is a descriptor that may run a query
when read; in Go the field is memory that was filled once by `Scan`. Same
syntax, different machinery — one can perform I/O, the other cannot.

**2.** `joinedload` on a collection multiplies rows: a parent with 10 children
arrives 10 times, and every parent column is repeated in each of those rows.
`selectinload` keeps each row once at the cost of one extra round-trip, which
is usually the better trade for one-to-many. For many-to-one, `joinedload` (or
a plain JOIN) is fine — no multiplication happens.

**3.** Because a JOIN duplicates the parent's columns for every child row, and
for one-to-many it changes the row count, forcing you to de-duplicate in
application code. With wide parent rows (long descriptions) the duplication is
also real bandwidth. One round-trip is not automatically cheaper than two.

**4.** N+1 costs one network round-trip per extra query. On localhost that is
tens of microseconds and hides inside the noise; across hosts it is a
millisecond or more, so 100 extra queries turn into an extra 100+ ms per
request. The query *count* was always wrong — only the per-query cost changed.

**5.** The JOIN multiplies rows: a product with three images comes back three
times, with the product columns repeated. `Scan` still returns one row at a
time, so the loop can no longer append blindly — it must detect when the
product id repeats, append the image to the product it already built, and only
start a new `Product` when the id changes. That bookkeeping is the reason the
two-query shape is usually preferred here.

---

Next: `GET /api/products` — the list endpoint, where the same JOIN meets
pagination, an optional filter and hand-written parameter validation.
