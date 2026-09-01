# ishopster

A learning e-commerce project — an SEO-friendly storefront.

**Stack:** Next.js (App Router, TypeScript) · FastAPI · SQLAlchemy 2 + Alembic ·
PostgreSQL · Docker Compose

```
Browser → Next.js (rendering, routing, SEO) → FastAPI (API, business logic)
        → SQLAlchemy → PostgreSQL
```

## Layout

| Path        | What                                                        |
| ----------- | ---------------------------------------------------------- |
| `backend/`  | FastAPI app, models, Alembic migrations                    |
| `frontend/` | Next.js app (catalog pages, metadata, sitemap, robots)     |
| `sql/`      | `seed.sql` (dev data) and `exercises.sql`                  |

## Run the stack

```bash
cp .env.example .env          # DB credentials, edit if needed
docker compose up -d --build
```

- Frontend: http://localhost:3000
- API + docs: http://localhost:8000 · http://localhost:8000/docs
- Postgres: `localhost:5432` (published for local tools)

```bash
docker compose ps             # db should be "healthy"
docker compose logs -f backend
docker compose down           # stop; keeps the data volume
docker compose down -v        # stop and delete the data volume
```

## Database

```bash
# apply migrations
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
# backend
cd backend && uv sync && uv run fastapi dev app/main.py

# frontend
cd frontend && pnpm install && pnpm dev
```

Configuration comes from the repo-root `.env` (backend) and
`frontend/.env.local` (frontend); both fall back to sensible localhost
defaults.
