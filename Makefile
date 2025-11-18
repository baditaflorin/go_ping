.PHONY: help build run test clean docker-build docker-run docker-compose-up docker-compose-down

# Variables
BINARY_NAME=ping
GO=go
GO_FLAGS=-v
DOCKER_IMAGE=ping:latest
DOCKER_REGISTRY?=ghcr.io/baditaflorin

help:
	@echo "Available targets:"
	@echo "  build              Build the application binary"
	@echo "  run                Run the application locally"
	@echo "  test               Run all tests"
	@echo "  docker-build       Build Docker image locally"
	@echo "  docker-run         Run Docker container"
	@echo "  docker-compose-up  Start services with docker-compose (includes Prometheus)"
	@echo "  docker-compose-down Stop docker-compose services"
	@echo "  clean              Clean build artifacts"
	@echo "  help               Show this help message"

## Build targets
build:
	@echo "Building $(BINARY_NAME)..."
	GOSUMDB=off CGO_ENABLED=0 GOOS=linux go build $(GO_FLAGS) -o bin/$(BINARY_NAME) .
	@echo "✓ Build complete: bin/$(BINARY_NAME)"

run:
	@echo "Running $(BINARY_NAME) with observability..."
	GOSUMDB=off $(GO) run $(GO_FLAGS) main.go

test:
	@echo "Running tests..."
	GOSUMDB=off $(GO) test $(GO_FLAGS) -race -cover ./...
	@echo "✓ Tests complete"

## Docker targets
docker-build:
	@echo "Building Docker image: $(DOCKER_IMAGE)"
	docker build -t $(DOCKER_IMAGE) .
	@echo "✓ Docker image built: $(DOCKER_IMAGE)"

docker-run: docker-build
	@echo "Starting Docker container..."
	docker run -d \
		--name $(BINARY_NAME) \
		--restart unless-stopped \
		-e PORT=8080 \
		-p 8080:8080 \
		$(DOCKER_IMAGE)
	@echo "✓ Container started. Access at http://localhost:8080"
	@echo "  - Pong: http://localhost:8080/"
	@echo "  - Health: http://localhost:8080/health"
	@echo "  - Metrics: http://localhost:8080/metrics"

docker-compose-up:
	@echo "Starting services with docker-compose..."
	docker-compose up -d --build
	@echo "✓ Services started:"
	@echo "  - Ping: http://localhost:8080"
	@echo "  - Health: http://localhost:8080/health"
	@echo "  - Metrics: http://localhost:8080/metrics"
	@echo "  - Prometheus: http://localhost:9090"

docker-compose-down:
	@echo "Stopping docker-compose services..."
	docker-compose down
	@echo "✓ Services stopped"

## Cleanup targets
clean:
	@echo "Cleaning up..."
	rm -f bin/$(BINARY_NAME)
	docker rm -f $(BINARY_NAME) 2>/dev/null || true
	docker rmi -f $(DOCKER_IMAGE) 2>/dev/null || true
	@echo "✓ Cleanup complete"

## Docker multi-arch build (requires docker buildx)
docker-buildx:
	@echo "Building multi-arch image: $(DOCKER_REGISTRY)/$(BINARY_NAME):latest"
	docker buildx build \
		--platform linux/amd64,linux/arm64 \
		-t $(DOCKER_REGISTRY)/$(BINARY_NAME):latest \
		--push .
	@echo "✓ Multi-arch build complete and pushed"
