# ishopster

A learning e-commerce project: Next.js + FastAPI + PostgreSQL.

## Milestone 1 — project foundation

For now only the database runs in Docker.

### Start the database

```bash
cp .env.example .env      # edit the values if needed
docker compose up -d
docker compose ps         # status should be "healthy"
```

### Connect to the database

```bash
docker compose exec db psql -U "$POSTGRES_USER" -d "$POSTGRES_DB"
```

From the host the database is available on `localhost:5432` (the port is
published for developer tools).

### Stop

```bash
docker compose down       # removes the container, keeps the data volume
docker compose down -v     # also removes the volume — data is lost
```
