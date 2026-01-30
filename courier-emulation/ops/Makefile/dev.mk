# DEVELOPMENT TASKS ====================================================================================================

##@ Development

.PHONY: run
run: ## Run the service locally
	go run ./cmd/courier-emulation

.PHONY: build
build: ## Build the binary
	go build -o bin/courier-emulation ./cmd/courier-emulation

.PHONY: test
test: ## Run tests
	go test ./... -v

.PHONY: test-coverage
test-coverage: ## Run tests with coverage
	go test ./... -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html

##@ Code Quality

.PHONY: lint
lint: ## Run linter
	golangci-lint run ./...

.PHONY: fmt
fmt: ## Format code
	go fmt ./...
	goimports -w .

##@ Code Generation

.PHONY: generate
generate: ## Run go generate (wire, etc.)
	go generate ./...

.PHONY: wire
wire: ## Generate wire_gen.go
	cd internal/di && wire

##@ Docker

.PHONY: docker-build
docker-build: ## Build Docker image
	docker build -f ops/dockerfile/Dockerfile -t courier-emulation:latest ../..

.PHONY: docker-run
docker-run: ## Run Docker container
	docker run --rm -it \
		-e OSRM_URL=http://host.docker.internal:5000 \
		-e KAFKA_BROKERS=host.docker.internal:9092 \
		-p 9090:9090 \
		courier-emulation:latest
