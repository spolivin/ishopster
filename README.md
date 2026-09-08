# ishopster

A learning e-commerce project — an SEO-friendly storefront.

**Stack:** Next.js (App Router, TypeScript) · FastAPI · SQLAlchemy 2 + Alembic ·
PostgreSQL · Docker Compose · a second API in Go (`net/http` + `pgx`)

```
Browser → Next.js (rendering, routing, SEO) → API (business logic) → PostgreSQL
```

The API exists in two interchangeable implementations that serve the same
contract, selected by a compose profile:

| Profile  | Service      | Implementation                                  |
| -------- | ------------ | ----------------------------------------------- |
| `python` | `backend`    | FastAPI + SQLAlchemy (the reference)             |
| `go`     | `backend-go` | Go standard library `net/http` + `pgx`, no ORM   |

Both answer on port 8000 with identical JSON, so the frontend cannot tell them
apart. The Go service is a learning exercise, not a replacement: **Alembic in
`backend/` remains the single owner of the database schema.**

## Layout

| Path         | What                                                       |
| ------------ | ---------------------------------------------------------- |
| `backend/`   | FastAPI app, models, Alembic migrations                    |
| `backend-go/`| Go implementation of the same read-only API                |
| `frontend/`  | Next.js app (catalog pages, metadata, sitemap, robots)     |
| `sql/`       | `seed.sql` (dev data) and `exercises.sql`                  |
| `learn/`     | Lesson notes written along the way                         |

## Run the stack

```bash
cp .env.example .env                              # DB credentials, edit if needed
docker compose --profile python up -d --build     # FastAPI backend
# or
docker compose --profile go up -d --build         # Go backend
```

A profile is required: both API services claim port 8000, so they can never
run together. Without one, `docker compose up` starts the database only. To
pick a default, set `COMPOSE_PROFILES=python` in `.env`.

- Frontend: http://localhost:3000
- API: http://localhost:8000 (FastAPI also serves http://localhost:8000/docs)
- Postgres: `localhost:5432` (published for local tools)

```bash
docker compose ps                       # db should be "healthy"
docker compose logs -f backend-go       # or backend, per profile
docker compose --profile go down        # stop; keeps the data volume
docker compose --profile go down -v     # stop and delete the data volume
```

Switching profiles needs no other change: `backend-go` also answers to the
network name `backend`, which is what the frontend fetches.

## Database

Schema changes always go through Alembic in `backend/`, which means the
`python` profile (the Go image is a bare binary — no shell, no Alembic). Run
migrations there first, then switch profiles if you want.

```bash
# apply migrations (python profile)
docker compose exec backend alembic upgrade head

# load sample data
docker compose exec -T db psql -U dbuser -d dbname < sql/seed.sql

# open a shell
docker compose exec db psql -U dbuser -d dbname
```

New migration after changing models in `backend/app/models.py`:

```bash
cd backend && uv run alembic revision --autogenerate -m "describe change"
```

## Local development (outside Docker)

```bash
# backend (FastAPI)
cd backend && uv sync && uv run fastapi dev app/main.py

# backend (Go) — reads POSTGRES_* from the environment, so export .env first
set -a; . ./.env; set +a
cd backend-go && go run .

# frontend
cd frontend && pnpm install && pnpm dev
```

Configuration comes from the repo-root `.env` (backends) and
`frontend/.env.local` (frontend); both fall back to sensible localhost
defaults. Unlike the FastAPI service, the Go one reads no `.env` file of its
own — it refuses to start unless `POSTGRES_USER`, `POSTGRES_PASSWORD` and
`POSTGRES_DB` are present in the environment.

Checks:

```bash
make backend-lint          # ruff
make backend-go-lint       # gofmt + go vet
make frontend-lint         # eslint
```
