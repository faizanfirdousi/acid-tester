.PHONY: up down run reset psql build all help

## Start the Postgres Docker container in the background
up:
	docker compose up -d
	@echo "⏳ Waiting for Postgres to be ready..."
	@sleep 3
	@docker inspect --format='Status: {{.State.Health.Status}}' acid_postgres

## Stop and remove the container
down:
	docker compose down

## Wipe all data (removes the Docker volume too)
reset:
	docker compose down -v

## Run the ACID tester
run:
	go run .

## Build a standalone binary
build:
	go build -o acid-tester .

## Connect to Postgres directly (interactive psql shell)
psql:
	docker exec -it acid_postgres psql -U acid_user -d acid_db

## One-shot: start container + run tests
all: up run

## Show this help
help:
	@echo ""
	@echo "  acid-tester — PostgreSQL ACID property tester"
	@echo ""
	@echo "  make up      Start Postgres Docker container"
	@echo "  make run     Run the ACID test suite"
	@echo "  make all     Start container + run tests"
	@echo "  make psql    Open interactive psql shell"
	@echo "  make down    Stop the container"
	@echo "  make reset   Wipe all data (removes volume)"
	@echo ""
