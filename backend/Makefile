.PHONY: build test clean run dev

# Build the application
build:
	go build -o bin/server ./cmd/server

# Run tests
test:
	go test ./...

# Run tests with coverage
test-coverage:
	go test -cover ./...

# Run tests with verbose output
test-verbose:
	go test -v ./...

# Clean build artifacts
clean:
	rm -rf bin/

# Run the application
run: build
	./bin/server

# Run in development mode (with hot reload using air if available)
dev:
	@which air > /dev/null && air || go run ./cmd/server

# Format code
fmt:
	go fmt ./...

# Lint code (requires golangci-lint)
lint:
	@which golangci-lint > /dev/null && golangci-lint run || echo "golangci-lint not installed, run: make install-linter"

# Install golangci-lint
install-linter:
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $$(go env GOPATH)/bin v1.54.2

# Run all checks (CI pipeline)
ci: fmt lint test build

# Pre-commit hook
pre-commit: fmt lint test

# Tidy dependencies
tidy:
	go mod tidy

# Download dependencies
deps:
	go mod download

# Build for production (with optimizations)
build-prod:
	CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags '-w -s' -o bin/server ./cmd/server