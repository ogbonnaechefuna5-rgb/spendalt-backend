include .env
export

.PHONY: run build migrate test clean

run:
	go run cmd/api/main.go

build:
	go build -o bin/api cmd/api/main.go

migrate:
	psql $(DATABASE_URL) -f migrations/001_init.sql

test:
	go test ./... -v

clean:
	rm -rf bin/

dev:
	air

docker-up:
	docker-compose up -d

docker-down:
	docker-compose down
