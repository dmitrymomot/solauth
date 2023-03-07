.PHONY: build tests

build:
	@echo "Building..."
	@go clean -cache
	@go build -o bin/server -v ./cmd/
	@echo "Build complete."

tests:
	@echo "Running tests..."
	@go test -v -timeout 300s -p 1 -count=1 -race -cover ./...
	@echo "Tests complete."