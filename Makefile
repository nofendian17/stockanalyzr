# Tooling
GO ?= go
PROTOC ?= protoc
GOBIN ?= $(HOME)/go/bin

# Paths
SERVICES_DIR := services
PKG_DIR := pkg
PROTO_DIR := api
PROTO_OUT_DIR := $(PKG_DIR)/gen

# Derived values
SERVICES := $(wildcard $(SERVICES_DIR)/*)
SERVICE_NAMES := $(notdir $(SERVICES))
SERVICE_PATH := $(SERVICES_DIR)/$(SERVICE)
PROTO_FILES := $(wildcard $(PROTO_DIR)/*.proto)

.PHONY: help deps proto mockgen build test tidy check-service run migrate migrate-up migrate-down migrate-up-all migrate-down-all run-% migrate-up-% migrate-down-% run-user migrate-up-user migrate-down-user

# --- Help ---
help:
	@echo "Available targets:"
	@echo "  deps                     Install required developer tools"
	@echo "  proto                    Generate protobuf files"
	@echo "  mockgen                  Run go generate for each service domain"
	@echo "  build                    Build all services"
	@echo "  test                     Run tests for all services"
	@echo "  tidy                     Tidy go modules in pkg and all services"
	@echo "  run SERVICE=<name>       Run a service with 'go run . serve'"
	@echo "  migrate-up SERVICE=<n>   Run up migration in selected service"
	@echo "  migrate-down SERVICE=<n> Run down migration in selected service"
	@echo "  migrate-up-all           Run up migrations for all services"
	@echo "  migrate-down-all         Run down migrations for all services"
	@echo "  run-<service>            Shortcut for run SERVICE=<service>"
	@echo "  migrate-up-<service>     Shortcut for migrate-up SERVICE=<service>"
	@echo "  migrate-down-<service>   Shortcut for migrate-down SERVICE=<service>"
	@echo ""
	@echo "Discovered services: $(SERVICE_NAMES)"

# --- Dependencies ---
deps:
	@echo "Installing dependencies..."
	$(GO) install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	$(GO) install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	$(GO) install go.uber.org/mock/mockgen@latest
	$(GO) install github.com/golang-migrate/migrate/v4/cmd/migrate@latest
	@echo "Dependencies installed."

# --- Protobuf ---
proto:
	@echo "Generating protobuf files..."
	@mkdir -p $(PROTO_OUT_DIR)
	@PATH="$(GOBIN):$$PATH" $(PROTOC) -I $(PROTO_DIR) \
		--go_out=$(PROTO_OUT_DIR) --go_opt=paths=source_relative \
		--go-grpc_out=$(PROTO_OUT_DIR) --go-grpc_opt=paths=source_relative \
		$(PROTO_FILES)
	@echo "Protobuf generation complete."

# --- Mock Generation ---
mockgen:
	@echo "Generating mocks for all services..."
	@for service in $(SERVICES); do \
		echo "Generating mocks in $$service..."; \
		if [ -d "$$service/internal/domain" ]; then \
			(cd $$service && $(GO) generate ./internal/domain/...); \
		else \
			echo "No internal/domain in $$service, skipping mockgen."; \
		fi \
	done
	@echo "Mock generation complete."

# --- Build / Test / Tidy ---
build:
	@echo "Building all services..."
	@for service in $(SERVICES); do \
		echo "Building $$service..."; \
		(cd $$service && $(GO) build ./...); \
	done
	@echo "Build complete."

test:
	@echo "Running tests for all services..."
	@for service in $(SERVICES); do \
		echo "Testing $$service..."; \
		(cd $$service && $(GO) test -v -count=1 ./...); \
	done
	@echo "Tests complete."

tidy:
	@echo "Tidying go modules..."
	@(cd $(PKG_DIR) && $(GO) mod tidy)
	@for service in $(SERVICES); do \
		echo "Tidying $$service..."; \
		(cd $$service && $(GO) mod tidy); \
	done
	@echo "Modules tidied."

# --- Service Selection Guard ---
check-service:
	@if [ -z "$(SERVICE)" ]; then \
		echo "Usage: make <target> SERVICE=<service-name>"; \
		echo "Available services: $(SERVICE_NAMES)"; \
		exit 1; \
	fi
	@if [ ! -d "$(SERVICE_PATH)" ]; then \
		echo "Service '$(SERVICE)' not found in $(SERVICES_DIR)/"; \
		echo "Available services: $(SERVICE_NAMES)"; \
		exit 1; \
	fi

# --- Run / Migrate ---
run: check-service
	@cd $(SERVICE_PATH) && $(GO) run . serve

migrate: check-service
	@cd $(SERVICE_PATH) && $(GO) run . migrate $(DIRECTION)

migrate-up: DIRECTION := up
migrate-up: migrate

migrate-down: DIRECTION := down
migrate-down: migrate

migrate-up-all:
	@for service in $(SERVICE_NAMES); do \
		echo "Migrating up $$service..."; \
		$(MAKE) migrate-up SERVICE=$$service || exit 1; \
	done

migrate-down-all:
	@for service in $(SERVICE_NAMES); do \
		echo "Migrating down $$service..."; \
		$(MAKE) migrate-down SERVICE=$$service || exit 1; \
	done

# Pattern shortcuts:
#   make run-user-service
#   make migrate-up-user-service
#   make migrate-down-user-service
run-%:
	@$(MAKE) run SERVICE=$*

migrate-up-%:
	@$(MAKE) migrate-up SERVICE=$*

migrate-down-%:
	@$(MAKE) migrate-down SERVICE=$*

# Backward-compatible aliases
run-user:
	@$(MAKE) run SERVICE=user-service

migrate-up-user:
	@$(MAKE) migrate-up SERVICE=user-service

migrate-down-user:
	@$(MAKE) migrate-down SERVICE=user-service

# Explicit per-service targets generated from discovered service directories.
define SERVICE_TARGETS
.PHONY: run-$(1) migrate-up-$(1) migrate-down-$(1)
run-$(1):
	@$(MAKE) run SERVICE=$(1)
migrate-up-$(1):
	@$(MAKE) migrate-up SERVICE=$(1)
migrate-down-$(1):
	@$(MAKE) migrate-down SERVICE=$(1)
endef

$(foreach svc,$(SERVICE_NAMES),$(eval $(call SERVICE_TARGETS,$(svc))))
