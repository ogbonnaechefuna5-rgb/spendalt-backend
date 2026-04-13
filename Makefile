include .env
export

.PHONY: run build migrate-up migrate-down migrate-reset migrate-status test clean dev docker-up docker-down

run:
	go run cmd/api/main.go

build:
	go build -o bin/api cmd/api/main.go

migrate-up:
	go run cmd/migrate/main.go up

migrate-down:
	go run cmd/migrate/main.go down

migrate-reset:
	go run cmd/migrate/main.go reset

migrate-status:
	go run cmd/migrate/main.go status

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
