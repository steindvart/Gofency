.PHONY: help db-up db-down db-logs db-shell dev-up dev-down build run docker-build docker-up docker-down docker-logs docker-restart

# Available commands help
help:
	@echo "Available commands:"
	@echo ""
	@echo "Local development:"
	@echo "  build         - Build the telegram-bot"
	@echo "  run           - Run the telegram-bot locally"
	@echo "  run-with-db   - Run the telegram-bot with PostgreSQL"
	@echo ""
	@echo "Docker commands:"
	@echo "  docker-build  - Build Docker image for bot"
	@echo "  docker-up     - Start all services (PostgreSQL + Bot)"
	@echo "  docker-down   - Stop all services"
	@echo "  docker-logs   - Show bot logs"
	@echo "  docker-restart- Restart bot container"
	@echo ""
	@echo "Database management:"
	@echo "  db-up         - Start PostgreSQL only"
	@echo "  db-down       - Stop PostgreSQL"
	@echo "  dev-up        - Start PostgreSQL + pgAdmin"
	@echo "  dev-down      - Stop all development services"

# Build and run application locally
build:
	go build -v -o bin/telegram-bot.exe ./cmd/bot

run:
	go run ./cmd/telegrambot/main.go

run-with-db: db-up
	go run ./cmd/telegrambot/main.go

# Docker commands
docker-build:
	docker-compose build telegrambot

docker-up:
	docker-compose up -d

docker-down:
	docker-compose down

docker-logs:
	docker-compose logs -f telegrambot

docker-restart:
	docker-compose restart telegrambot

# Database management
db-up:
	docker-compose up -d postgres

db-down:
	docker-compose down

# Development with pgAdmin
dev-up:
	docker-compose --profile dev up -d

dev-down:
	docker-compose --profile dev down
