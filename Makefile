# Variables
POSTGRES_CONTAINER=$(shell docker ps -qf "name=postgres")
KAFKA_CONTAINER=$(shell docker ps -qf "name=kafka")

.PHONY: help up down worker cron api produce-order db-check

help: ## Show this help message
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

up: ## Start infrastructure (DB, Redis, Kafka)
	docker-compose up -d

down: ## Stop all infrastructure
	docker-compose down

db-check: ## View current user metrics in Postgres
	docker exec -it $(POSTGRES_CONTAINER) psql -U user -d daffodil -c "SELECT * FROM user_metrics;"

worker: ## Run the Kafka Consumer Worker
	go run cmd/worker/main.go

cron: ## Run the Segment Evaluator (Cron Job)
	go run cmd/cron/main.go

api: ## Run the Experiment API
	go run cmd/api/main.go

produce-order: ## Send a mock order for User U1 to Kafka
	@echo '{"user_id": "U1", "amount": 500.0}' | docker exec -i $(KAFKA_CONTAINER) /opt/kafka/bin/kafka-console-producer.sh --bootstrap-server 127.0.0.1:9092 --topic order_events
	@echo "Sent order event for U1"

seed-data: ## Pump 30 orders for User U1 to trigger Power User status
	@echo "Pushing 30 orders to Kafka..."
	@for i in {1..30}; do \
		echo '{"user_id": "U1", "amount": 150.0}' | docker exec -i $(shell docker ps -qf "name=kafka") /opt/kafka/bin/kafka-console-producer.sh --bootstrap-server 127.0.0.1:9092 --topic order_events; \
	done
	@echo "✅ Done. Now run 'make cron' to update segments."

# Run everything for development
dev-backend:
	go run cmd/api/main.go

dev-frontend:
	cd dashboard && npm run dev

dev-worker:
	go run cmd/worker/main.go

# Install all dependencies at once
install:
	go mod download
	cd dashboard && npm install

# Start the infrastructure
infra:
	docker-compose up -d