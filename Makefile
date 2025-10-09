# ----- Makefile (Phase 0) -----
SHELL := bash
APP_NAME := rate-limiter
CMD_DIR := ./cmd/api
PKG := ./...
GO := go

# Default envs (overridable). Tip: usually HTTP_ADDRESS is like ":1234"
HTTP_ADDRESS ?= :1234
RATE_LIMIT   ?= 100
WINDOW_SEC   ?= 0
REDIS_URL    ?= :9001

# Build output
BIN_DIR := bin
BIN := $(BIN_DIR)/$(APP_NAME)

.PHONY: help
help: ## Show available targets
	@awk 'BEGIN {FS = ":.*##"; printf "\n\033[1m%s\033[0m\n\n", "Available targets"} /^[a-zA-Z0-9_\-]+:.*?##/ { printf "  \033[36m%-18s\033[0m %s\n", $$1, $$2 }' $(MAKEFILE_LIST)

.PHONY: run
run: ## Run the API with env vars
	@HTTP_ADDRESS="$(HTTP_ADDRESS)" RATE_LIMIT="$(RATE_LIMIT)" WINDOW_SEC="$(WINDOW_SEC)" REDIS_URL="$(REDIS_URL)" \
		$(GO) run $(CMD_DIR)

.PHONY: run-env
run-env: ## Run using .env if present (bash 'set -a' auto-exports)
	@set -a; [ -f .env ] && source .env; set +a; \
		$(GO) run $(CMD_DIR)

.PHONY: build
build: ## Build binary to ./bin
	@mkdir -p $(BIN_DIR)
	@HTTP_ADDRESS="$(HTTP_ADDRESS)" RATE_LIMIT="$(RATE_LIMIT)" WINDOW_SEC="$(WINDOW_SEC)" REDIS_URL="$(REDIS_URL)" \
		$(GO) build -o $(BIN) $(CMD_DIR)
	@echo "Built $(BIN)"

.PHONY: tidy
tidy: ## go mod tidy
	$(GO) mod tidy

.PHONY: fmt
fmt: ## go fmt
	$(GO) fmt $(PKG)

.PHONY: vet
vet: ## go vet
	$(GO) vet $(PKG)

.PHONY: test
test: ## Run tests
	$(GO) test -count=1 $(PKG)

.PHONY: test-race
test-race: ## Run tests with race detector
	$(GO) test -race -count=1 $(PKG)

.PHONY: lint
lint: ## Run golangci-lint if installed
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run ./...; \
	else \
		echo "golangci-lint not found (skip). Install: https://golangci-lint.run/"; \
	fi

.PHONY: clean
clean: ## Remove build artifacts
	@rm -rf $(BIN_DIR)

.PHONY: env
env: ## Show effective env vars
	@echo HTTP_ADDRESS=$(HTTP_ADDRESS)
	@echo RATE_LIMIT=$(RATE_LIMIT)
	@echo WINDOW_SEC=$(WINDOW_SEC)
	@echo REDIS_URL=$(REDIS_URL)
