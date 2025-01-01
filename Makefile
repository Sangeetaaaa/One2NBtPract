# Variables
DB_CONTAINER_NAME=mongodb
API_IMAGE_NAME=student-api
NETWORK_NAME=app-network

# Check if MongoDB is running
check-db:
	@if [ -z "$$(docker ps -q -f name=$(DB_CONTAINER_NAME))" ]; then \
		echo "MongoDB is not running. Starting MongoDB..." && \
		$(MAKE) db-start; \
	else \
		echo "MongoDB is already running."; \
	fi

# Check if migrations are applied
check-migrations:
	@echo "Checking if migrations are applied..."
	@if ! docker exec $(DB_CONTAINER_NAME) mongosh --eval \
		'db = db.getSiblingDB("school"); \
		if (db.students.exists()) { exit(0) } else { exit(1) }' >/dev/null 2>&1; then \
		echo "Migrations not found. Applying migrations..." && \
		$(MAKE) db-migrate; \
	else \
		echo "Migrations already applied."; \
	fi

# Start MongoDB container
db-start:
	@echo "Starting MongoDB container..."
	@docker network create $(NETWORK_NAME) 2>/dev/null || true
	@if [ -z "$$(docker ps -a -q -f name=$(DB_CONTAINER_NAME))" ]; then \
		docker run --name $(DB_CONTAINER_NAME) \
			--network $(NETWORK_NAME) \
			-v mongodb_data:/data/db \
			-p 27017:27017 \
			-d mongo:latest && \
		echo "Waiting for MongoDB to start..." && \
		sleep 5; \
	else \
		docker start $(DB_CONTAINER_NAME); \
	fi

# Run database migrations
db-migrate:
	@echo "Running database migrations..."
	@docker exec $(DB_CONTAINER_NAME) mongosh --eval \
		'db = db.getSiblingDB("school"); \
		db.createCollection("students"); \
		if (!db.students.indexExists("email_1")) { \
			db.students.createIndex({"email": 1}, {unique: true}); \
		}'

# Start the API using docker-compose
api-start: check-db check-migrations
	@echo "Starting API container using docker-compose..."
	docker-compose up -d api

# Stop everything
stop:
	@echo "Stopping all containers..."
	docker-compose down
	@if [ ! -z "$$(docker ps -q -f name=$(DB_CONTAINER_NAME))" ]; then \
		docker stop $(DB_CONTAINER_NAME); \
	fi

# Clean everything
clean: stop
	@echo "Cleaning up containers and volumes..."
	docker-compose down -v
	@if [ ! -z "$$(docker ps -a -q -f name=$(DB_CONTAINER_NAME))" ]; then \
		docker rm $(DB_CONTAINER_NAME); \
	fi
	docker volume rm mongodb_data 2>/dev/null || true
	docker network rm $(NETWORK_NAME) 2>/dev/null || true

# Build the API Docker image
build:
	@echo "Building API Docker image..."
    sudo docker-compose build --progress=plain api

# Run tests for the API (assumes API has a test command in package.json)
test:
	@echo "Running tests..."
	sudo docker-compose run --rm api npm test

# Perform linting (assumes API has a lint command in package.json)
lint:
	@echo "Running linting..."
	sudo docker-compose run --rm api npm run lint

# Help target
help:
	@echo "Available targets:"
	@echo "  api-start    - Start the complete application (DB + API)"
	@echo "  db-start     - Start MongoDB container only"
	@echo "  db-migrate   - Run database migrations only"
	@echo "  stop         - Stop all containers"
	@echo "  clean        - Remove all containers, volumes, and networks"
	@echo "  build        - Build the API Docker image"
	@echo "  test         - Run tests"
	@echo "  lint         - Run linting"
	@echo "  help         - Show this help message"

.PHONY: check-db check-migrations db-start db-migrate api-start stop clean build test lint help
