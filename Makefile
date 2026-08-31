backend-lint:
	cd backend && uv run ruff check .

backend-format:
	cd backend && uv run ruff check . --fix && uv run ruff format .

sql-exercises:
	docker compose exec -T db psql -U dbuser -d dbname < sql/exercises.sql