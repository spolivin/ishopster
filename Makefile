backend-lint:
	cd backend && uv run ruff check .

backend-format:
	cd backend && uv run ruff check . --fix && uv run ruff format .

frontend-lint:
	cd frontend && pnpm lint

frontend-typecheck:
	cd frontend && pnpm typecheck

frontend-format:
	cd frontend && pnpm format

frontend-format-check:
	cd frontend && pnpm format:check

sql-exercises:
	docker compose exec -T db psql -U dbuser -d dbname < sql/exercises.sql
