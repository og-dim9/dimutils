# Variables
APP_NAME := dimutils
VERSION := 0.3.0
IMAGE_NAME := dimutils
IMAGE_TAG := $(VERSION)
REGISTRY := docker.io/dim9

# OS Detection
ifeq ($(OS),Windows_NT)
    detected_OS := windows
    EXT := .exe
else
    detected_OS := $(shell uname | tr A-Z a-z)
    EXT :=
endif

# Go compiler
GO := go
LDFLAGS := -ldflags "-X main.Version=$(VERSION)"

# Individual tool packages (now integrated into multicall binary)
TOOL_NAMES := cbxxml2regex consume ebcdic eventdiff gitaskop mkgchat produce regex2json serve tandum togchat unexpect

# Build directories
BUILD_DIR := build
BIN_DIR := $(BUILD_DIR)/bin

# Build targets
.PHONY: all clean test docker docker-push docker-run multicall individual $(TOOL_NAMES)

default: $(detected_OS)

all: multicall individual

# Download dependencies
dl:
	@echo "Downloading dependencies..."
	@$(GO) mod download

# Create build directories
$(BIN_DIR):
	@mkdir -p $(BIN_DIR)

# Build multicall binary
multicall: dl $(BIN_DIR)
	@echo "Building multicall binary..."
	@GOOS=$(detected_OS) $(GO) build $(LDFLAGS) -o $(BIN_DIR)/$(APP_NAME)$(EXT) ./cmd/$(APP_NAME)

# Build multicall for Linux
linux: dl $(BIN_DIR)
	@echo "Building multicall binary for Linux..."
	@GOOS=linux $(GO) build $(LDFLAGS) -o $(BIN_DIR)/$(APP_NAME) ./cmd/$(APP_NAME)

# Build multicall for Windows
windows: dl $(BIN_DIR)
	@echo "Building multicall binary for Windows..."
	@GOOS=windows GOARCH=amd64 $(GO) build $(LDFLAGS) -o $(BIN_DIR)/$(APP_NAME).exe ./cmd/$(APP_NAME)

# Create symlinks for individual tools (all tools are now in multicall binary)
individual: symlinks

# Create symlinks for multicall (Unix only)
symlinks: multicall
ifeq ($(detected_OS),windows)
	@echo "Symlinks not supported on Windows"
else
	@echo "Creating symlinks for multicall binary..."
	@cd $(BIN_DIR) && for tool in $(TOOL_NAMES); do \
		ln -sf $(APP_NAME) $$tool; \
	done
endif

# Test targets
test:
	@echo "Running tests..."
	@$(GO) test ./...

test-tools:
	@echo "All tools are now integrated into multicall binary. Use 'make test' instead."

# Docker targets
docker: linux
	@echo "Building Docker image..."
	@docker build -t $(IMAGE_NAME):$(IMAGE_TAG) .

docker-tag: docker
	@echo "Tagging Docker image..."
	@docker tag $(IMAGE_NAME):$(IMAGE_TAG) $(REGISTRY)/$(IMAGE_NAME):$(IMAGE_TAG)
	@docker tag $(IMAGE_NAME):$(IMAGE_TAG) $(REGISTRY)/$(IMAGE_NAME):latest

docker-push: docker-tag
	@echo "Pushing Docker image..."
	@docker push $(REGISTRY)/$(IMAGE_NAME):$(IMAGE_TAG)
	@docker push $(REGISTRY)/$(IMAGE_NAME):latest

docker-run: docker
	@docker run -it --rm $(IMAGE_NAME):$(IMAGE_TAG)

# Run targets
run: multicall
	@$(BIN_DIR)/$(APP_NAME)$(EXT)

run-linux: linux
	@$(BIN_DIR)/$(APP_NAME)

run-windows: windows
	@$(BIN_DIR)/$(APP_NAME).exe

# Clean targets
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf $(BUILD_DIR)

# Development targets
watch:
	@echo "Watching for changes..."
	@fswatch -o . | xargs -n1 -I{} make run

# Help target
help:
	@echo "Available targets:"
	@echo "  all          - Build multicall and individual binaries"
	@echo "  multicall    - Build multicall binary for current OS"
	@echo "  individual   - Build all individual tool binaries"
	@echo "  linux        - Build multicall binary for Linux"
	@echo "  windows      - Build multicall binary for Windows"
	@echo "  symlinks     - Create symlinks for multicall (Unix only)"
	@echo "  test         - Run Go tests"
	@echo "  test-tools   - Run tool-specific tests"
	@echo "  docker       - Build Docker image"
	@echo "  docker-push  - Build and push Docker image"
	@echo "  clean        - Clean all build artifacts"
	@echo "  run          - Build and run multicall binary"
	@echo "  watch        - Watch for changes and rebuild"
	@echo ""
	@echo "Individual tools: $(TOOL_NAMES)"
