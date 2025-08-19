# Orion Platform Makefile

.PHONY: build run test clean help

# Build the operator
build:
	@echo "🔨 Building operator..."
	go build -o bin/operator ./cmd/operator

# Run the operator locally  
run: build
	@echo "🚀 Running operator..."
	./bin/operator

# Run tests
test:
	@echo "🧪 Running tests..."
	go test ./...

# Clean build artifacts
clean:
	@echo "🧹 Cleaning..."
	rm -rf bin/

# Show help
help:
	@echo "Available commands:"
	@echo "  build  - Build the operator binary"
	@echo "  run    - Build and run the operator"
	@echo "  test   - Run all tests"
	@echo "  clean  - Clean build artifacts"
