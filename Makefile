.PHONY: up down test migrate-up migrate-down

up:
	docker compose up --build

down:
	docker compose down

test:
	go test ./...

migrate-up:
	docker compose run --rm migrate up

migrate-down:
	docker compose run --rm migrate down
