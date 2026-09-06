backend-lint:
	cd backend && uv run ruff check .

backend-format:
	cd backend && uv run ruff check . --fix && uv run ruff format .

# `gofmt -l` only lists unformatted files, it exits 0 either way, so the
# non-empty list has to be turned into a failure by hand.
backend-go-lint:
	cd backend-go && unformatted=$$(gofmt -l .); \
	  if [ -n "$$unformatted" ]; then echo "not gofmt'd:"; echo "$$unformatted"; exit 1; fi
	cd backend-go && go vet ./...

backend-go-format:
	cd backend-go && gofmt -w .

frontend-lint:
	cd frontend && pnpm lint

frontend-typecheck:
	cd frontend && pnpm typecheck

frontend-format:
	cd frontend && pnpm format

frontend-format-check:
	cd frontend && pnpm format:check

sql-seed:
	docker compose exec -T db psql -U dbuser -d dbname < sql/seed.sql
