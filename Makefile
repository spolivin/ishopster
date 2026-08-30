backend-lint:
	cd backend && uv run ruff check .

backend-format:
	cd backend && uv run ruff check . --fix && uv run ruff format .