.PHONY: up down build rebuild logs shell clean stop restart status install-compose build-frontend build-backend build-cli dev-cli test fmt lint install check install-linter tidy deps dev preview clean-build ci pre-commit build-prod build-release release help deploy-user remove-user create-role remove-role deploy-stackset status-stackset delete-stackset cli-status check-aws-config

# Docker compose command (try docker-compose first, fallback to docker compose)
DOCKER_COMPOSE := $(shell command -v docker-compose 2> /dev/null || echo "docker compose")

# Default target
help:
	@echo "AWS IAM Manager - Unified Build System"
	@echo ""
	@echo "🏗️  Build Targets:"
	@echo "  build-frontend   - Build Vue.js frontend"
	@echo "  build-backend    - Build Go backend server"
	@echo "  build-cli        - Build Go CLI application"
	@echo "  build-go         - Build both backend and CLI"
	@echo "  build            - Build everything (Docker)"
	@echo "  build-prod       - Production build with optimizations"
	@echo "  build-release    - Multi-platform release binaries"
	@echo ""
	@echo "🚀 Development:"
	@echo "  dev              - Run Docker development environment"
	@echo "  dev-cli          - Run CLI in development mode"
	@echo "  dev-frontend     - Run frontend development server"
	@echo "  dev-backend      - Run backend in development mode"
	@echo ""
	@echo "🧪 Testing & Quality:"
	@echo "  test             - Run all tests"
	@echo "  test-coverage    - Run tests with coverage"
	@echo "  fmt              - Format all code"
	@echo "  lint             - Lint all code"
	@echo "  check            - Run all checks (fmt + lint + test)"
	@echo "  ci               - CI pipeline (install + check + build)"
	@echo ""
	@echo "🐳 Docker Operations:"
	@echo "  up               - Start services"
	@echo "  down             - Stop services"
	@echo "  logs             - Show logs"
	@echo "  shell            - Access container shell"
	@echo "  restart          - Restart services"
	@echo "  status           - Show service status"
	@echo ""
	@echo "☁️  AWS IAM Management:"
	@echo "  deploy-user      - Deploy IAM user and resources"
	@echo "  remove-user      - Remove IAM user and resources"
	@echo "  create-role      - Create IAM role for cross-account access"
	@echo "  remove-role      - Remove IAM role and resources"
	@echo "  deploy-stackset  - Deploy StackSet for organization setup"
	@echo "  status-stackset  - Show StackSet deployment status"
	@echo "  delete-stackset  - Delete StackSet and all instances"
	@echo "  cli-status       - Show current deployment status"
	@echo ""
	@echo "🧹 Cleanup:"
	@echo "  clean            - Clean everything"
	@echo "  clean-build      - Clean build artifacts only"
	@echo ""
	@echo "🔧 Setup & Configuration:"
	@echo "  check-aws-config - Verify AWS credentials and configuration"
	@echo ""
	@echo "📖 CLI Usage Examples:"
	@echo "  cli-help         - Show CLI help"

# ============================================================================
# BUILD TARGETS
# ============================================================================

# Frontend build
build-frontend:
	@echo "📦 Building frontend..."
	cd frontend && npm install && npm run build

# Backend build
build-backend:
	@echo "🔧 Building backend server..."
	cd backend && go build -o ../bin/aws-iam-manager ./cmd/server

# CLI build
build-cli:
	@echo "⚙️  Building CLI..."
	@if command -v go >/dev/null 2>&1; then \
		mkdir -p bin; \
		go mod tidy; \
		go mod download; \
		CGO_ENABLED=0 go build -ldflags="-w -s" -o bin/iam-manager ./cmd/iam-manager; \
	else \
		echo "❌ Error: Go is not installed. Please install Go 1.21+"; \
		exit 1; \
	fi

# Build Go projects (backend + CLI)
build-go: build-backend build-cli

# Build everything with Docker
build:
	@echo "🐳 Building Docker services..."
	$(DOCKER_COMPOSE) build

# Production build with optimizations
build-prod:
	@echo "🚀 Building for production..."
	$(MAKE) build-frontend
	cd backend && CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags '-w -s' -o ../bin/aws-iam-manager-prod ./cmd/server
	$(MAKE) build-cli

# Multi-platform release binaries
build-release:
	@echo "📦 Building release binaries for multiple platforms..."
	@mkdir -p bin/release
	$(MAKE) build-frontend
	# Backend server
	cd backend && GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o ../bin/release/aws-iam-manager-server-linux-amd64 ./cmd/server
	cd backend && GOOS=linux GOARCH=arm64 go build -ldflags="-w -s" -o ../bin/release/aws-iam-manager-server-linux-arm64 ./cmd/server
	cd backend && GOOS=darwin GOARCH=amd64 go build -ldflags="-w -s" -o ../bin/release/aws-iam-manager-server-darwin-amd64 ./cmd/server
	cd backend && GOOS=darwin GOARCH=arm64 go build -ldflags="-w -s" -o ../bin/release/aws-iam-manager-server-darwin-arm64 ./cmd/server
	cd backend && GOOS=windows GOARCH=amd64 go build -ldflags="-w -s" -o ../bin/release/aws-iam-manager-server-windows-amd64.exe ./cmd/server
	# CLI binary
	GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o bin/release/iam-manager-linux-amd64 ./cmd/iam-manager
	GOOS=linux GOARCH=arm64 go build -ldflags="-w -s" -o bin/release/iam-manager-linux-arm64 ./cmd/iam-manager
	GOOS=darwin GOARCH=amd64 go build -ldflags="-w -s" -o bin/release/iam-manager-darwin-amd64 ./cmd/iam-manager
	GOOS=darwin GOARCH=arm64 go build -ldflags="-w -s" -o bin/release/iam-manager-darwin-arm64 ./cmd/iam-manager
	GOOS=windows GOARCH=amd64 go build -ldflags="-w -s" -o bin/release/iam-manager-windows-amd64.exe ./cmd/iam-manager
	@echo "✅ Release binaries built in bin/release/ directory"

# ============================================================================
# DEVELOPMENT TARGETS
# ============================================================================

# Docker development environment
dev: rebuild logs

# Rebuild Docker environment
rebuild: down build up

# Frontend development server
dev-frontend:
	@echo "🎨 Starting frontend development server..."
	cd frontend && npm run dev

# Backend development server
dev-backend:
	@echo "🔧 Starting backend development server..."
	cd backend && @which air > /dev/null && air || go run ./cmd/server

# CLI development
dev-cli:
	@if command -v go >/dev/null 2>&1; then \
		go run ./cmd/iam-manager; \
	else \
		echo "❌ Error: Go is not installed. Cannot run dev-cli target."; \
		exit 1; \
	fi

# Preview production frontend build
preview:
	@echo "👀 Previewing frontend production build..."
	cd frontend && npm run preview

# ============================================================================
# TESTING & QUALITY TARGETS
# ============================================================================

# Run all tests
test:
	@echo "🧪 Running tests..."
	cd backend && go test ./...
	go test ./cmd/iam-manager
	cd frontend && npm test

# Run tests with coverage
test-coverage:
	@echo "🧪 Running tests with coverage..."
	cd backend && go test -cover ./...
	cd frontend && npm run test -- --coverage

# Run tests with verbose output
test-verbose:
	@echo "🧪 Running verbose tests..."
	cd backend && go test -v ./...

# Format all code
fmt:
	@echo "✨ Formatting code..."
	@if command -v go >/dev/null 2>&1; then \
		echo "  📝 Formatting Go code..."; \
		go fmt ./cmd/iam-manager/...; \
		cd backend && go fmt ./...; \
	else \
		echo "⚠️  Go not installed, skipping Go formatting"; \
	fi
	@echo "  📝 Formatting frontend code..."
	@cd frontend && (which prettier > /dev/null && npm run format || echo "    Prettier not configured")

# Lint all code
lint:
	@echo "🔍 Linting code..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		echo "  🔍 Running golangci-lint on Go code..."; \
		golangci-lint run ./cmd/iam-manager/...; \
		cd backend && golangci-lint run; \
	else \
		echo "  ⚠️  golangci-lint not found, using basic go vet"; \
		go vet ./cmd/iam-manager; \
		cd backend && go vet ./...; \
	fi
	@echo "  🔍 Linting frontend code..."
	@cd frontend && (which eslint > /dev/null && npm run lint || echo "    ESLint not configured")

# Run all checks (pre-commit)
check: fmt lint test
	@echo "✅ All checks passed!"

# CI pipeline
ci: tidy deps check build
	@echo "✅ CI pipeline completed successfully!"

# Pre-commit checks (lighter than CI)
pre-commit: fmt lint test

# ============================================================================
# DEPENDENCY MANAGEMENT
# ============================================================================

# Tidy Go dependencies
tidy:
	@echo "🧹 Tidying Go dependencies..."
	go mod tidy
	cd backend && go mod tidy

# Download Go dependencies
deps:
	@echo "📦 Downloading dependencies..."
	go mod download
	cd backend && go mod download
	cd frontend && npm install

# Install frontend dependencies (production)
install-frontend:
	@echo "📦 Installing frontend dependencies..."
	cd frontend && npm ci

# ============================================================================
# DOCKER OPERATIONS
# ============================================================================

# Start services
up:
	@echo "🚀 Starting Docker services..."
	$(DOCKER_COMPOSE) up -d

# Stop services
down:
	@echo "🛑 Stopping Docker services..."
	$(DOCKER_COMPOSE) down

# Show logs
logs:
	@echo "📋 Showing service logs..."
	$(DOCKER_COMPOSE) logs -f

# Show logs for specific service
logs-service:
	$(DOCKER_COMPOSE) logs -f aws-iam-manager

# Access shell in running container
shell:
	$(DOCKER_COMPOSE) exec aws-iam-manager sh

# Stop services without removing containers
stop:
	$(DOCKER_COMPOSE) stop

# Restart services
restart:
	$(DOCKER_COMPOSE) restart

# Show status of services
status:
	$(DOCKER_COMPOSE) ps

# ============================================================================
# CLI USAGE EXAMPLES
# ============================================================================

cli-help: build-cli
	@echo "📖 Showing CLI help..."
	@if [ -f bin/iam-manager ]; then \
		./bin/iam-manager --help; \
	elif command -v go >/dev/null 2>&1; then \
		go run ./cmd/iam-manager --help; \
	else \
		echo "❌ Error: Neither binary nor Go found. Run 'make build-cli' first."; \
		exit 1; \
	fi

# ============================================================================
# SETUP & CONFIGURATION TARGETS
# ============================================================================

# Check AWS credentials and configuration
check-aws-config:
	@echo "🔍 Checking AWS configuration..."
	@echo ""
	@echo "📋 AWS CLI Configuration:"
	@aws configure list || echo "❌ AWS CLI not configured"
	@echo ""
	@echo "🌍 Environment Variables:"
	@echo "  AWS_REGION: $${AWS_REGION:-<not set>}"
	@echo "  AWS_ACCESS_KEY_ID: $${AWS_ACCESS_KEY_ID:-<not set>}"
	@echo "  AWS_SECRET_ACCESS_KEY: $${AWS_SECRET_ACCESS_KEY:+<set>}$${AWS_SECRET_ACCESS_KEY:-<not set>}"
	@echo ""
	@echo "💡 Quick Setup Guide:"
	@echo "   1. Configure AWS CLI: aws configure"
	@echo "   2. Or set environment variables:"
	@echo "      export AWS_ACCESS_KEY_ID=your_key"
	@echo "      export AWS_SECRET_ACCESS_KEY=your_secret" 
	@echo "      export AWS_REGION=us-east-1"
	@echo "   3. Or copy .env.example to .env and set your credentials"
	@echo ""
	@echo "🧪 Testing AWS connectivity..."
	@if command -v aws >/dev/null 2>&1; then \
		aws sts get-caller-identity 2>/dev/null || echo "❌ AWS connectivity test failed - credentials may be invalid"; \
	else \
		echo "⚠️  AWS CLI not installed - install it for easier credential management"; \
	fi

# ============================================================================
# AWS IAM MANAGEMENT TARGETS
# ============================================================================

# Deploy IAM user and resources
deploy-user: build-cli
	@echo "🚀 Deploying IAM user and resources..."
	@if [ -f bin/iam-manager ]; then \
		./bin/iam-manager deploy; \
	elif command -v go >/dev/null 2>&1; then \
		go run ./cmd/iam-manager deploy; \
	else \
		echo "❌ Error: Neither binary nor Go found. Run 'make build-cli' first."; \
		exit 1; \
	fi

# Remove IAM user and resources
remove-user: build-cli
	@echo "🗑️  Removing IAM user and resources..."
	@if [ -f bin/iam-manager ]; then \
		./bin/iam-manager remove; \
	elif command -v go >/dev/null 2>&1; then \
		go run ./cmd/iam-manager remove; \
	else \
		echo "❌ Error: Neither binary nor Go found. Run 'make build-cli' first."; \
		exit 1; \
	fi

# Create IAM role for cross-account access
create-role: build-cli
	@echo "🔐 Creating IAM role for cross-account access..."
	@if [ -f bin/iam-manager ]; then \
		./bin/iam-manager create-role; \
	elif command -v go >/dev/null 2>&1; then \
		go run ./cmd/iam-manager create-role; \
	else \
		echo "❌ Error: Neither binary nor Go found. Run 'make build-cli' first."; \
		exit 1; \
	fi

# Remove IAM role and resources
remove-role: build-cli
	@echo "🗑️  Removing IAM role and resources..."
	@if [ -f bin/iam-manager ]; then \
		./bin/iam-manager remove-role; \
	elif command -v go >/dev/null 2>&1; then \
		go run ./cmd/iam-manager remove-role; \
	else \
		echo "❌ Error: Neither binary nor Go found. Run 'make build-cli' first."; \
		exit 1; \
	fi

# Deploy StackSet for organization setup
deploy-stackset: build-cli
	@echo "📦 Deploying StackSet for organization setup..."
	@if [ -f bin/iam-manager ]; then \
		./bin/iam-manager stackset-deploy; \
	elif command -v go >/dev/null 2>&1; then \
		go run ./cmd/iam-manager stackset-deploy; \
	else \
		echo "❌ Error: Neither binary nor Go found. Run 'make build-cli' first."; \
		exit 1; \
	fi

# Show StackSet deployment status
status-stackset: build-cli
	@echo "📊 Checking StackSet deployment status..."
	@if [ -f bin/iam-manager ]; then \
		./bin/iam-manager stackset-status; \
	elif command -v go >/dev/null 2>&1; then \
		go run ./cmd/iam-manager stackset-status; \
	else \
		echo "❌ Error: Neither binary nor Go found. Run 'make build-cli' first."; \
		exit 1; \
	fi

# Delete StackSet and all instances
delete-stackset: build-cli
	@echo "🗑️  Deleting StackSet and all instances..."
	@if [ -f bin/iam-manager ]; then \
		./bin/iam-manager stackset-delete; \
	elif command -v go >/dev/null 2>&1; then \
		go run ./cmd/iam-manager stackset-delete; \
	else \
		echo "❌ Error: Neither binary nor Go found. Run 'make build-cli' first."; \
		exit 1; \
	fi

# Show current deployment status
cli-status: build-cli
	@echo "📋 Showing current deployment status..."
	@if [ -f bin/iam-manager ]; then \
		./bin/iam-manager status; \
	elif command -v go >/dev/null 2>&1; then \
		go run ./cmd/iam-manager status; \
	else \
		echo "❌ Error: Neither binary nor Go found. Run 'make build-cli' first."; \
		exit 1; \
	fi

# ============================================================================
# CLEANUP TARGETS
# ============================================================================

# Clean everything (Docker + build artifacts)
clean:
	@echo "🧹 Cleaning everything..."
	$(DOCKER_COMPOSE) down -v --rmi all --remove-orphans || true
	rm -rf bin/
	rm -rf frontend/dist/
	rm -rf frontend/node_modules/
	cd backend && rm -rf bin/

# Clean only build artifacts
clean-build:
	@echo "🧹 Cleaning build artifacts..."
	rm -rf bin/
	rm -rf frontend/dist/
	cd backend && rm -rf bin/

# ============================================================================
# INSTALLATION TARGETS
# ============================================================================

# Install CLI globally
install:
	@echo "📦 Installing CLI globally..."
	go install ./cmd/iam-manager

# Install golangci-lint
install-linter:
	@echo "📦 Installing golangci-lint..."
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $$(go env GOPATH)/bin v1.54.2
	@echo "✅ golangci-lint installed to $$(go env GOPATH)/bin"

# Install docker-compose (requires sudo)
install-compose:
	@echo "📦 Installing docker-compose to /usr/local/bin (requires sudo)..."
	sudo curl -L "https://github.com/docker/compose/releases/latest/download/docker-compose-$(shell uname -s)-$(shell uname -m)" -o /usr/local/bin/docker-compose
	sudo chmod +x /usr/local/bin/docker-compose
	@echo "✅ docker-compose installed successfully"

# Install docker-compose to user bin (no sudo)
install-compose-local:
	@echo "📦 Installing docker-compose to ~/bin..."
	@mkdir -p ~/bin
	curl -L "https://github.com/docker/compose/releases/latest/download/docker-compose-$(shell uname -s)-$(shell uname -m)" -o ~/bin/docker-compose
	chmod +x ~/bin/docker-compose
	@echo "✅ docker-compose installed to ~/bin/docker-compose"
	@echo "💡 Add ~/bin to your PATH: export PATH=\$$HOME/bin:\$$PATH"