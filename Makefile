.PHONY: build buildMac buildWinX64 buildUnix tests

tests:
	@echo "Running tests..."
	@go test -v -timeout 300s -p 1 -count=1 -race -cover ./...
	@echo "Tests complete."

build: buildMac buildUnix

buildMac:
	@echo "Building for Mac..."
	@go clean -cache
	@GOOS=darwin GOARCH=amd64 go build -o bin/server.mac -v ./cmd/
	@echo "Build complete."

buildWinX64:
	@echo "Building for Windows x64..."
	@go clean -cache
	@GOOS=windows GOARCH=amd64 go build -o bin/server.windows.exe -v ./cmd/
	@echo "Build complete."

buildUnix:
	@echo "Building for Unix..."
	@go clean -cache
	@GOOS=linux GOARCH=amd64 go build -o bin/server.unix -v ./cmd/
	@echo "Build complete."