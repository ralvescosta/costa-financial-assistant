# ─── Costa Financial Assistant — Root Makefile ──────────────────────────────
#
# Targets:
#   dev-up                 Start all services and frontend in development mode
#   svc/run/<service>      Run a specific backend service
#   svc/test/<service>     Run unit tests for a specific backend service
#   migrate/up/<service>   Apply DB migrations for a specific service
#   migrate/down/<service> Rollback DB migrations for a specific service
#   proto/generate         Generate Go + gRPC code from .proto files
#   frontend/dev           Start the Vite development server
#   frontend/test          Run frontend Vitest hook tests
#   frontend/build         Build frontend for production

SHELL  := /bin/bash
GOROOT := $(shell go env GOROOT)
GOPATH := $(shell go env GOPATH)

# ─── Services ────────────────────────────────────────────────────────────────
SERVICES := bff bills files identity onboarding payments
MIGRATION_SERVICES := onboarding files bills identity payments

# ─── Colours ─────────────────────────────────────────────────────────────────
CYAN  := \033[0;36m
RESET := \033[0m

.PHONY: help dev-up frontend/dev frontend/test frontend/build proto/generate \
        $(addprefix svc/run/,$(SERVICES)) \
        $(addprefix svc/test/,$(SERVICES)) \
        $(addprefix migrate/up/,$(SERVICES)) \
	$(addprefix migrate/down/,$(SERVICES)) \
	migrate/up migrate/down migrate/status migrate/validate \
	local dev stg prd

help: ## Show this help
	@grep -E '^[a-zA-Z/_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
	  awk 'BEGIN {FS = ":.*?## "}; {printf "$(CYAN)%-30s$(RESET) %s\n", $$1, $$2}'

# ─── Development bootstrap ───────────────────────────────────────────────────
dev-up: ## Start infrastructure + all backend services + frontend in dev mode
	@docker compose --profile dev up -d
	@make -j2 frontend/dev svc/run/bff

# ─── Frontend ────────────────────────────────────────────────────────────────
frontend/dev: ## Start Vite dev server
	@cd frontend && npm run dev

frontend/test: ## Run Vitest hook unit tests
	@cd frontend && npm run test

frontend/build: ## Build frontend for production
	@cd frontend && npm run build

# ─── Service name → Viper env-var prefix mapping ─────────────────────────────
SERVICE_PREFIX_bff        := BFF
SERVICE_PREFIX_bills      := BILLS
SERVICE_PREFIX_files      := FILES
SERVICE_PREFIX_identity   := IDENTITY
SERVICE_PREFIX_onboarding := ONBOARDING
SERVICE_PREFIX_payments   := PAYMENTS

# ─── Default per-service bind ports (avoid 9090 collisions) ─────────────────
HTTP_PORT_bff       ?= 8080
GRPC_PORT_identity  ?= 9091
GRPC_PORT_files     ?= 9092
GRPC_PORT_bills     ?= 9093
GRPC_PORT_onboarding?= 9094
GRPC_PORT_payments  ?= 9095

# ─── Backend service targets ─────────────────────────────────────────────────
define SERVICE_TARGETS
svc/run/$(1): ## Run backend service: $(1)
	@cd backend && \
	  $(SERVICE_PREFIX_$(1))_SERVICE_NAME=$(1) \
	  $(SERVICE_PREFIX_$(1))_DB_DSN=$(DB_URL_$(1)) \
	  $(SERVICE_PREFIX_$(1))_HTTP_PORT=$(HTTP_PORT_$(1)) \
	  $(SERVICE_PREFIX_$(1))_GRPC_PORT=$(GRPC_PORT_$(1)) \
	  go run . $(1)

svc/test/$(1): ## Run unit tests for backend service: $(1)
	@cd backend && go test -race -count=1 ./internals/$(1)/...

endef
$(foreach svc,$(SERVICES),$(eval $(call SERVICE_TARGETS,$(svc))))

# ─── Integration tests ───────────────────────────────────────────────────────
test/integration: ## Run backend integration tests with ephemeral DB
	@cd backend && go test -race -count=1 -v -tags integration ./tests/integration/...

# ─── Migrations ──────────────────────────────────────────────────────────────
DB_URL_bff         ?= postgres://postgres:postgres@localhost:5432/financial_payments?sslmode=disable
DB_URL_onboarding  ?= postgres://postgres:postgres@localhost:5432/financial_onboarding?sslmode=disable
DB_URL_files       ?= postgres://postgres:postgres@localhost:5432/financial_files?sslmode=disable
DB_URL_bills       ?= postgres://postgres:postgres@localhost:5432/financial_bills?sslmode=disable
DB_URL_payments    ?= postgres://postgres:postgres@localhost:5432/financial_payments?sslmode=disable

MIGRATIONS_DB_DSN ?= postgres://postgres:postgres@localhost:5432/financial_assistant?sslmode=disable
MIGRATIONS_ENV    ?= local

# Supports invocation like: make migrate/up --env local
CLI_ENV := $(firstword $(filter local dev stg prd,$(MAKECMDGOALS)))
ifneq ($(CLI_ENV),)
MIGRATIONS_ENV := $(CLI_ENV)
endif

local dev stg prd:
	@:

migrate/up: ## Apply migrations for all services
	@cd backend && \
	  MIGRATIONS_SERVICE_NAME=migrations \
	  MIGRATIONS_DB_DSN=$(MIGRATIONS_DB_DSN) \
	  go run . migrations up --env $(MIGRATIONS_ENV)

migrate/down: ## Rollback one migration for all services
	@cd backend && \
	  for svc in $(MIGRATION_SERVICES); do \
	    MIGRATIONS_SERVICE_NAME=migrations \
	    MIGRATIONS_DB_DSN=$(MIGRATIONS_DB_DSN) \
	    go run . migrations down --service $$svc --env $(MIGRATIONS_ENV); \
	  done

migrate/status: ## Show migration status
	@cd backend && \
	  MIGRATIONS_SERVICE_NAME=migrations \
	  MIGRATIONS_DB_DSN=$(MIGRATIONS_DB_DSN) \
	  go run . migrations status --format table

migrate/validate: ## Validate migration folders and file pairs
	@cd backend && go run . migrations validate

define MIGRATE_TARGETS
migrate/up/$(1): ## Apply migrations for service: $(1)
	@cd backend && \
	  MIGRATIONS_SERVICE_NAME=migrations \
	  MIGRATIONS_DB_DSN=$(MIGRATIONS_DB_DSN) \
	  go run . migrations up --service $(1) --env $(MIGRATIONS_ENV)

migrate/down/$(1): ## Rollback one migration for service: $(1)
	@cd backend && \
	  MIGRATIONS_SERVICE_NAME=migrations \
	  MIGRATIONS_DB_DSN=$(MIGRATIONS_DB_DSN) \
	  go run . migrations down --service $(1) --env $(MIGRATIONS_ENV)

endef
$(foreach svc,$(MIGRATION_SERVICES),$(eval $(call MIGRATE_TARGETS,$(svc))))

# ─── Proto generation ────────────────────────────────────────────────────────
PROTO_SRC_DIR := backend/protos
PROTO_GEN_DIR := backend/protos/generated
PROTO_MODULES := common/v1 onboarding/v1 identity/v1 files/v1 bills/v1 payments/v1

PROTOC_GEN_GO      := $(shell go env GOBIN)/protoc-gen-go
PROTOC_GEN_GO_GRPC := $(shell go env GOBIN)/protoc-gen-go-grpc

proto/generate: ## Regenerate Go + gRPC code from all .proto files
	@mkdir -p $(PROTO_GEN_DIR)
	@for module in $(PROTO_MODULES); do \
	  echo "Generating $$module …"; \
	  protoc \
	    --proto_path=$(PROTO_SRC_DIR) \
	    --plugin=protoc-gen-go=$(PROTOC_GEN_GO) \
	    --go_out=$(PROTO_GEN_DIR) \
	    --go_opt=paths=source_relative \
	    --plugin=protoc-gen-go-grpc=$(PROTOC_GEN_GO_GRPC) \
	    --go-grpc_out=$(PROTO_GEN_DIR) \
	    --go-grpc_opt=paths=source_relative \
	    $(PROTO_SRC_DIR)/$$module/*.proto; \
	done
	@echo "Proto generation complete."
