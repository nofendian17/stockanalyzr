# Directories
SERVICES_DIR := services
PROTO_DIR := api
PROTO_OUT_DIR := pkg/gen

# Find all service directories (any subdirectory inside services/)
SERVICES := $(wildcard $(SERVICES_DIR)/*)

.PHONY: deps proto build test mockgen $(SERVICES)

# --- Dependencies ---
deps:
	@echo "Installing dependencies..."
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	go install go.uber.org/mock/mockgen@latest
	go install github.com/golang-migrate/migrate/v4/cmd/migrate@latest
	@echo "Dependencies installed."

# --- Protobuf ---
proto:
	@echo "Generating protobuf files..."
	@mkdir -p $(PROTO_OUT_DIR)
	@PATH="$(HOME)/go/bin:$$PATH" protoc -I $(PROTO_DIR) \
		--go_out=$(PROTO_OUT_DIR) --go_opt=paths=source_relative \
		--go-grpc_out=$(PROTO_OUT_DIR) --go-grpc_opt=paths=source_relative \
		$(wildcard $(PROTO_DIR)/*.proto)
	@echo "Protobuf generation complete."

# --- Mock Generation ---
# This will go through every service and run go generate
mockgen:
	@echo "Generating mocks for all services..."
	@for service in $(SERVICES); do \
		echo "Generating mocks in $$service..."; \
		if [ -d "$$service/internal/domain" ]; then \
			cd $$service && go generate ./internal/domain/... && cd ../..; \
		else \
			echo "No internal/domain in $$service, skipping mockgen."; \
		fi \
	done
	@echo "Mock generation complete."

# --- Build ---
# Builds all services
build:
	@echo "Building all services..."
	@for service in $(SERVICES); do \
		echo "Building $$service..."; \
		cd $$service && go build ./... && cd ../..; \
	done
	@echo "Build complete."

# --- Test ---
# Tests all services
test:
	@echo "Running tests for all services..."
	@for service in $(SERVICES); do \
		echo "Testing $$service..."; \
		cd $$service && go test -v -count=1 ./... && cd ../..; \
	done
	@echo "Tests complete."

# --- Tidy ---
# Tidies mod files for pkg and all services
tidy:
	@echo "Tidying go modules..."
	@cd pkg && go mod tidy && cd ..
	@for service in $(SERVICES); do \
		echo "Tidying $$service..."; \
		cd $$service && go mod tidy && cd ../..; \
	done
	@echo "Modules tidied."

# --- Run User Service (Convenience) ---
run-user:
	cd services/user-service && go run . serve

migrate-up-user:
	cd services/user-service && go run . migrate up

migrate-down-user:
	cd services/user-service && go run . migrate down
