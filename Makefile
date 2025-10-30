.PHONY: build install test clean dev adapter-test help

# Build the binary
build:
	@echo "Building deploy-tunnel..."
	@go build -o dt ./cmd/deploy-tunnel
	@echo "✓ Build complete: ./dt"

# Install to /usr/local/bin
install: build
	@echo "Installing dt to /usr/local/bin..."
	@sudo mv dt /usr/local/bin/
	@echo "✓ Installed! Run 'dt help' to get started"

# Run tests
test:
	@echo "Running Go tests..."
	@go test -v ./...
	@echo "Running adapter tests..."
	@cd adapters && bun test

# Run Go tests with coverage
test-coverage:
	@go test -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "✓ Coverage report: coverage.html"

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@rm -f dt
	@rm -f coverage.out coverage.html
	@echo "✓ Clean complete"

# Development build (with race detector)
dev:
	@echo "Building with race detector..."
	@go build -race -o dt ./cmd/deploy-tunnel
	@echo "✓ Development build complete"

# Test a specific adapter
adapter-test:
	@echo "Testing Vercel adapter..."
	@echo '{}' | bun run adapters/vercel/index.ts capabilities
	@echo "✓ Adapter test complete"

# Install Go dependencies
deps:
	@echo "Installing Go dependencies..."
	@go mod download
	@cd adapters && bun install
	@echo "✓ Dependencies installed"

# Format code
fmt:
	@echo "Formatting code..."
	@go fmt ./...
	@cd adapters && bun fmt
	@echo "✓ Code formatted"

# Lint code
lint:
	@echo "Linting Go code..."
	@go vet ./...
	@echo "✓ Lint complete"

# Show help
help:
	@echo "Deploy Tunnel - Makefile Commands"
	@echo ""
	@echo "  make build          Build the dt binary"
	@echo "  make install        Install dt to /usr/local/bin"
	@echo "  make test           Run all tests"
	@echo "  make test-coverage  Generate test coverage report"
	@echo "  make clean          Remove build artifacts"
	@echo "  make dev            Build with race detector"
	@echo "  make adapter-test   Test adapter communication"
	@echo "  make deps           Install dependencies"
	@echo "  make fmt            Format code"
	@echo "  make lint           Lint code"
	@echo "  make help           Show this help message"

# Default target
all: fmt lint build test
