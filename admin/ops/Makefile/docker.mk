# DOCKER TASKS =========================================================================================================

COMPOSE_DIR := $(SELF_DIR)../..

##@ Docker

.PHONY: up
up: ## Run services in development mode
	COMPOSE_PROFILES=dns,observability,gateway docker compose \
		-f $(COMPOSE_DIR)/docker-compose.yaml \
		up -d --remove-orphans --build

.PHONY: down
down: confirm ## Stop and remove containers
	COMPOSE_PROFILES=dns,observability,gateway docker compose \
		-f $(COMPOSE_DIR)/docker-compose.yaml \
		down --remove-orphans
	docker network prune -f

.PHONY: logs
logs: ## Show container logs
	docker compose -f $(COMPOSE_DIR)/docker-compose.yaml logs -f

.PHONY: build
build: ## Build Docker image
	docker build -f ops/dockerfile/Dockerfile -t shop-admin .
