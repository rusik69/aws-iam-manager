.PHONY: up down build rebuild logs shell clean stop restart status install-compose

# Docker compose command (try docker-compose first, fallback to docker compose)
DOCKER_COMPOSE := $(shell command -v docker-compose 2> /dev/null || echo "docker compose")

# Start services
up:
	$(DOCKER_COMPOSE) up -d

# Stop services
down:
	$(DOCKER_COMPOSE) down

# Build services
build:
	$(DOCKER_COMPOSE) build

# Rebuild and start services
rebuild: down build up

# Show logs
logs:
	$(DOCKER_COMPOSE) logs -f

# Show logs for specific service
logs-service:
	$(DOCKER_COMPOSE) logs -f aws-iam-manager

# Access shell in running container
shell:
	$(DOCKER_COMPOSE) exec aws-iam-manager sh

# Clean up (remove containers, networks, volumes, and images)
clean:
	$(DOCKER_COMPOSE) down -v --rmi all --remove-orphans

# Stop services without removing containers
stop:
	$(DOCKER_COMPOSE) stop

# Restart services
restart:
	$(DOCKER_COMPOSE) restart

# Show status of services
status:
	$(DOCKER_COMPOSE) ps

# Install docker-compose (standalone) - requires sudo
install-compose:
	@echo "Installing docker-compose to /usr/local/bin (requires sudo)..."
	sudo curl -L "https://github.com/docker/compose/releases/latest/download/docker-compose-$(shell uname -s)-$(shell uname -m)" -o /usr/local/bin/docker-compose
	sudo chmod +x /usr/local/bin/docker-compose
	@echo "docker-compose installed successfully"

# Install docker-compose to local user bin (no sudo required)
install-compose-local:
	@echo "Installing docker-compose to ~/bin..."
	@mkdir -p ~/bin
	curl -L "https://github.com/docker/compose/releases/latest/download/docker-compose-$(shell uname -s)-$(shell uname -m)" -o ~/bin/docker-compose
	chmod +x ~/bin/docker-compose
	@echo "docker-compose installed to ~/bin/docker-compose"
	@echo "Add ~/bin to your PATH if not already done: export PATH=\$$HOME/bin:\$$PATH"

# Development mode - build and run with logs
dev: rebuild logs