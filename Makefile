.PHONY: help proto build test docker-build docker-push k8s-deploy k8s-delete local-dev clean

# Variables
REGISTRY ?= your-registry
USER_SERVICE_IMAGE = $(REGISTRY)/go-drive-user-service
API_GATEWAY_IMAGE = $(REGISTRY)/go-drive-api-gateway
VERSION ?= latest

help: ## Display this help message
	@echo "Go-Drive Microservices Makefile"
	@echo ""
	@echo "Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-20s %s\n", $$1, $$2}'

proto: ## Generate protobuf code
	@echo "Generating protobuf code..."
	cd proto && make all

proto-install: ## Install protobuf tools
	@echo "Installing protobuf tools..."
	cd proto && make install-tools

build: proto ## Build all services
	@echo "Building user-service..."
	go build -o bin/user-service ./services/user-service
	@echo "Building api-gateway..."
	go build -o bin/api-gateway ./services/api-gateway
	@echo "Build complete!"

test: ## Run tests
	@echo "Running tests..."
	go test -v ./...

docker-build: ## Build Docker images
	@echo "Building Docker images..."
	docker build -t $(USER_SERVICE_IMAGE):$(VERSION) -f services/user-service/Dockerfile .
	docker build -t $(API_GATEWAY_IMAGE):$(VERSION) -f services/api-gateway/Dockerfile .
	@echo "Docker images built successfully!"

docker-push: docker-build ## Push Docker images to registry
	@echo "Pushing Docker images..."
	docker push $(USER_SERVICE_IMAGE):$(VERSION)
	docker push $(API_GATEWAY_IMAGE):$(VERSION)
	@echo "Docker images pushed successfully!"

docker-compose-up: ## Start services with docker-compose
	@echo "Starting services with docker-compose..."
	docker-compose -f docker-compose.microservices.yml up -d
	@echo "Services started. Check logs with: make docker-compose-logs"

docker-compose-down: ## Stop services with docker-compose
	@echo "Stopping services..."
	docker-compose -f docker-compose.microservices.yml down

docker-compose-logs: ## View docker-compose logs
	docker-compose -f docker-compose.microservices.yml logs -f

k8s-deploy: ## Deploy to Kubernetes
	@echo "Deploying to Kubernetes..."
	kubectl apply -k k8s/base
	@echo "Deployment complete. Check status with: make k8s-status"

k8s-delete: ## Delete Kubernetes deployment
	@echo "Deleting Kubernetes resources..."
	kubectl delete -k k8s/base
	@echo "Resources deleted."

k8s-status: ## Check Kubernetes deployment status
	@echo "Checking deployment status..."
	kubectl get all -n go-drive

k8s-logs-user: ## View user-service logs in Kubernetes
	kubectl logs -f -n go-drive -l app=user-service

k8s-logs-gateway: ## View api-gateway logs in Kubernetes
	kubectl logs -f -n go-drive -l app=api-gateway

k8s-port-forward: ## Port forward API gateway to localhost:8080
	@echo "Port forwarding api-gateway to localhost:8080..."
	kubectl port-forward -n go-drive service/api-gateway 8080:80

local-dev: ## Run services locally for development
	@echo "Starting user-service..."
	./bin/user-service &
	@echo "Starting api-gateway..."
	./bin/api-gateway &
	@echo "Services started locally. Press Ctrl+C to stop."

clean: ## Clean build artifacts
	@echo "Cleaning..."
	rm -rf bin/
	rm -f proto/user/*.pb.go
	rm -f proto/file/*.pb.go
	@echo "Clean complete!"

deps: ## Download Go dependencies
	@echo "Downloading dependencies..."
	go mod download
	go mod tidy

fmt: ## Format Go code
	@echo "Formatting code..."
	go fmt ./...

lint: ## Run linters
	@echo "Running linters..."
	golangci-lint run ./...

db-migrate: ## Run database migrations (requires Supabase CLI)
	@echo "Running database migrations..."
	supabase db push

# Testing
test-unit: ## Run unit tests
	go test -short -v ./...

test-integration: ## Run integration tests
	go test -v ./...

test-coverage: ## Generate test coverage report
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Docker operations
docker-clean: ## Remove Docker images
	docker rmi $(USER_SERVICE_IMAGE):$(VERSION) || true
	docker rmi $(API_GATEWAY_IMAGE):$(VERSION) || true

# Quick commands
quick-start: proto build docker-compose-up ## Quick start with docker-compose

quick-stop: docker-compose-down ## Quick stop docker-compose

rebuild: clean build ## Clean and rebuild

# API testing
test-api: ## Test API endpoints (requires running services)
	@echo "Testing health endpoint..."
	curl -f http://localhost:8080/health
	@echo "\nCreating test user..."
	curl -X POST http://localhost:8080/api/v1/users \
		-H "Content-Type: application/json" \
		-d '{"first_name":"Test","surname":"User","email":"test@example.com"}'
	@echo "\nListing users..."
	curl http://localhost:8080/api/v1/users
