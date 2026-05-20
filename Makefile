include .env
export

# ── Variables ─────────────────────────────────────────────────────────────────
DB_CONTAINER  := moninte-postgres
REDIS_CONTAINER := moninte-redis
DB_USER       := moninte
DB_NAME       := moninte
DB_PORT       := 5435

.PHONY: \
	run build dev clean \
	test test-race test-cover \
	migrate-up migrate-down migrate-reset migrate-status \
	docker-up docker-down docker-restart docker-logs docker-ps \
	docker-build docker-rebuild \
	db-shell db-clear db-drop db-dump db-restore db-size db-tables \
	redis-shell redis-clear redis-info \
	logs-api logs-db logs-redis

# ── App ───────────────────────────────────────────────────────────────────────

run:
	go run cmd/api/main.go

build:
	go build -o bin/api cmd/api/main.go

dev:
	air

clean:
	rm -rf bin/ coverage.out

# ── Tests ─────────────────────────────────────────────────────────────────────

test:
	go test ./tests/... -v

test-race:
	go test ./tests/... -race

test-cover:
	go test ./tests/... -coverprofile=coverage.out
	go tool cover -html=coverage.out

# ── Migrations ────────────────────────────────────────────────────────────────

migrate-up:
	go run cmd/migrate/main.go up

migrate-down:
	go run cmd/migrate/main.go down

migrate-reset:
	go run cmd/migrate/main.go reset

migrate-status:
	go run cmd/migrate/main.go status

# ── Docker — lifecycle ────────────────────────────────────────────────────────

docker-up:
	docker compose up -d

docker-down:
	docker compose down

docker-restart:
	docker compose restart

docker-ps:
	docker compose ps

docker-logs:
	docker compose logs -f

# ── Docker — build ────────────────────────────────────────────────────────────

docker-build:
	docker compose build

docker-rebuild:
	docker compose down
	docker compose build --no-cache
	docker compose up -d

# ── Docker — individual service logs ─────────────────────────────────────────

logs-api:
	docker compose logs -f api

logs-db:
	docker compose logs -f postgres

logs-redis:
	docker compose logs -f redis

# ── Database ──────────────────────────────────────────────────────────────────

## Open an interactive psql shell
db-shell:
	docker exec -it $(DB_CONTAINER) psql -U $(DB_USER) -d $(DB_NAME)

## Delete all rows from every table (keeps schema intact)
db-clear:
	@echo "Clearing all data from $(DB_NAME)..."
	docker exec -i $(DB_CONTAINER) psql -U $(DB_USER) -d $(DB_NAME) -c \
		"DO \$$\$$ DECLARE r RECORD; BEGIN FOR r IN (SELECT tablename FROM pg_tables WHERE schemaname = 'public') LOOP EXECUTE 'TRUNCATE TABLE ' || quote_ident(r.tablename) || ' CASCADE'; END LOOP; END \$$\$$;"
	@echo "Done."

## Drop and recreate the database (destructive — loses schema too)
db-drop:
	@echo "WARNING: This will destroy all data and schema in $(DB_NAME)."
	@read -p "Type 'yes' to confirm: " confirm && [ "$$confirm" = "yes" ] || (echo "Aborted." && exit 1)
	docker exec -i $(DB_CONTAINER) psql -U postgres -c "DROP DATABASE IF EXISTS $(DB_NAME);"
	docker exec -i $(DB_CONTAINER) psql -U postgres -c "CREATE DATABASE $(DB_NAME) OWNER $(DB_USER);"
	@echo "Database recreated. Run 'make migrate-up' to restore schema."

## Dump the database to a file (usage: make db-dump FILE=backup.sql)
FILE ?= backup_$(shell date +%Y%m%d_%H%M%S).sql
db-dump:
	docker exec -i $(DB_CONTAINER) pg_dump -U $(DB_USER) $(DB_NAME) > $(FILE)
	@echo "Dumped to $(FILE)"

## Restore the database from a file (usage: make db-restore FILE=backup.sql)
db-restore:
	@[ -f "$(FILE)" ] || (echo "FILE not set or not found. Usage: make db-restore FILE=backup.sql" && exit 1)
	docker exec -i $(DB_CONTAINER) psql -U $(DB_USER) -d $(DB_NAME) < $(FILE)
	@echo "Restored from $(FILE)"

## Show database size
db-size:
	docker exec -i $(DB_CONTAINER) psql -U $(DB_USER) -d $(DB_NAME) -c \
		"SELECT pg_size_pretty(pg_database_size('$(DB_NAME)')) AS size;"

## List all tables with row counts
db-tables:
	docker exec -i $(DB_CONTAINER) psql -U $(DB_USER) -d $(DB_NAME) -c \
		"SELECT relname AS table, n_live_tup AS rows FROM pg_stat_user_tables ORDER BY n_live_tup DESC;"

# ── Redis ─────────────────────────────────────────────────────────────────────

## Open an interactive redis-cli shell
redis-shell:
	docker exec -it $(REDIS_CONTAINER) redis-cli

## Flush all Redis keys (clears token store, rate limit counters)
redis-clear:
	docker exec -i $(REDIS_CONTAINER) redis-cli FLUSHALL
	@echo "Redis cleared."

## Show Redis memory and key stats
redis-info:
	docker exec -i $(REDIS_CONTAINER) redis-cli INFO server | grep -E "redis_version|uptime"
	docker exec -i $(REDIS_CONTAINER) redis-cli INFO memory | grep -E "used_memory_human|maxmemory_human"
	docker exec -i $(REDIS_CONTAINER) redis-cli INFO keyspace
