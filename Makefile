BINARY := bin/api

.PHONY: help run build test cover fmt vet tidy check clean mongo mongo-stop

help: ## Show available targets
	@grep -E '^[a-zA-Z_-]+:.*?## ' $(MAKEFILE_LIST) | awk -F':.*?## ' '{printf "  \033[36m%-12s\033[0m %s\n", $$1, $$2}'

dev: ## Run the API server
	@test -f .env || { cp .env.example .env; echo "created .env from .env.example - review it before running"; }
	go run ./cmd/api

build: ## Build the API binary into bin/
	go build -o $(BINARY) ./cmd/api

test: ## Run all tests
	go test ./...

cover: ## Run tests and open the HTML coverage report
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out

fmt: ## Format all Go files
	gofmt -w .

vet: ## Run go vet
	go vet ./...

tidy: ## Sync go.mod and go.sum
	go mod tidy

check: fmt vet test ## Format, vet and test

clean: ## Remove build artifacts
	rm -rf bin coverage.out

mongo: ## Start a local MongoDB in Docker
	docker run -d --name user-management-mongo -p 27017:27017 mongo:7

mongo-stop: ## Stop and remove the local MongoDB container
	docker rm -f user-management-mongo

# generate secret for JWT
secret:
	@openssl rand -base64 32
