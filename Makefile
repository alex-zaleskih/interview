.PHONY: mongo-up mongo-down mongo-reset seed run build test

## Start MongoDB in Docker (seed data is loaded on the first start)
mongo-up:
	docker compose up -d --wait mongo

## Stop MongoDB (data is kept)
mongo-down:
	docker compose down

## Recreate MongoDB from scratch with fresh seed data
mongo-reset:
	docker compose down -v
	docker compose up -d --wait mongo

## Reload seed data into the running MongoDB
seed:
	docker compose exec -T mongo mongosh --quiet /docker-entrypoint-initdb.d/01-seed.js

run:
	go run ./cmd/server

build:
	go build -o bin/server ./cmd/server

test:
	go test ./...
