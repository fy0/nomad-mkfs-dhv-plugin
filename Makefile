.PHONY: help build install install-scripts clean test

BINARY_NAME=mkfs
INSTALL_DIR=/usr/local/bin
SCRIPTS_DIR=scripts

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-20s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: ## Build the plugin binary
	@echo "Building $(BINARY_NAME)..."
	go build -o $(BINARY_NAME) ./cmd/mkfs

install: build ## Install the plugin to system directory
	@echo "Installing $(BINARY_NAME) to $(INSTALL_DIR)..."
	sudo install -m 755 $(BINARY_NAME) $(INSTALL_DIR)/$(BINARY_NAME)
	@echo "Installed successfully!"

install-scripts: ## Install helper scripts to system directory
	@echo "Installing helper scripts..."
	sudo install -m 755 $(SCRIPTS_DIR)/volume-usage.sh $(INSTALL_DIR)/nomad-dhv-usage
	sudo install -m 755 $(SCRIPTS_DIR)/volume-manager.sh $(INSTALL_DIR)/nomad-dhv-manager
	@echo "Scripts installed:"
	@echo "  - nomad-dhv-usage"
	@echo "  - nomad-dhv-manager"

test: ## Run tests
	go test -v ./...

clean: ## Clean build artifacts
	@echo "Cleaning..."
	rm -f $(BINARY_NAME)

# For development
dev-setup: ## Setup development environment
	@echo "Setting up development environment..."
	go mod download
	chmod +x $(SCRIPTS_DIR)/*.sh
	@echo "Development environment ready!"
