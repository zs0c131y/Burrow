.PHONY: build clean test install help

# Variables
BINARY_NAME=wm.exe
VERSION=1.0.0
BUILD_DIR=build
LDFLAGS=-ldflags="-s -w -X main.version=$(VERSION)"

help: ## Show this help message
	@echo "Burrow Build System"
	@echo ""
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-15s %s\n", $$1, $$2}'

build: ## Build Burrow for Windows
	@echo "Building Burrow..."
	@go build $(LDFLAGS) -o $(BINARY_NAME)
	@echo "Build complete: $(BINARY_NAME)"

build-linux: ## Cross-compile from Linux/Mac
	@echo "Cross-compiling for Windows..."
	@GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o $(BINARY_NAME)
	@echo "Build complete: $(BINARY_NAME)"

build-all: ## Build for multiple platforms
	@echo "Building for all platforms..."
	@mkdir -p $(BUILD_DIR)
	@GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/wm-windows-amd64.exe
	@GOOS=windows GOARCH=386 go build $(LDFLAGS) -o $(BUILD_DIR)/wm-windows-386.exe
	@echo "All builds complete in $(BUILD_DIR)/"

clean: ## Remove build artifacts
	@echo "Cleaning build artifacts..."
	@rm -f $(BINARY_NAME)
	@rm -rf $(BUILD_DIR)
	@echo "Clean complete"

deps: ## Download dependencies
	@echo "Downloading dependencies..."
	@go mod download
	@go mod tidy
	@echo "Dependencies updated"

test: ## Run tests
	@echo "Running tests..."
	@go test -v ./...

fmt: ## Format code
	@echo "Formatting code..."
	@go fmt ./...
	@echo "Format complete"

lint: ## Run linter
	@echo "Running linter..."
	@golangci-lint run
	@echo "Lint complete"

install: build ## Install to system (Windows)
	@echo "Installing to C:\Windows\System32..."
	@copy $(BINARY_NAME) C:\Windows\System32\
	@echo "Installation complete"

dev: ## Build and run
	@make build
	@./$(BINARY_NAME)

release: ## Create release build
	@echo "Creating release build v$(VERSION)..."
	@mkdir -p $(BUILD_DIR)
	@GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/wm-v$(VERSION)-windows-amd64.exe
	@echo "Release build complete: $(BUILD_DIR)/wm-v$(VERSION)-windows-amd64.exe"

run: build ## Build and run with args (use ARGS="...")
	@./$(BINARY_NAME) $(ARGS)
