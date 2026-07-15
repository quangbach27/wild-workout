SERVICES := common trainer training user

SERVICE ?=
TAG ?=

.PHONY: test lint help

help:
	@echo "Usage: make test SERVICE=<name> [TAG=<tag>]"
	@echo "       make lint SERVICE=<name>"
	@echo ""
	@echo "  SERVICE  one of: $(SERVICES)"
	@echo "  TAG      omitted        -> unit tests only"
	@echo "           component     -> component tests (spins up the service; needs postgres running)"
	@echo "           integration   -> integration tests against adapters/db (needs postgres running)"
	@echo ""
	@echo "Examples:"
	@echo "  make test SERVICE=trainer"
	@echo "  make test SERVICE=training TAG=component"
	@echo "  make test SERVICE=user TAG=integration"
	@echo "  make lint SERVICE=user"
	@echo ""
	@echo "postgres can be started with: docker compose up -d postgres"

test:
ifeq ($(strip $(SERVICE)),)
	$(error SERVICE is required, e.g. "make test SERVICE=trainer". Valid services: $(SERVICES))
endif
ifeq ($(filter $(SERVICE),$(SERVICES)),)
	$(error unknown SERVICE "$(SERVICE)". Valid services: $(SERVICES))
endif
ifeq ($(strip $(TAG)),)
	go test ./internal/$(SERVICE)/...
else
	go test -tags=$(TAG) ./internal/$(SERVICE)/...
endif

lint:
ifeq ($(strip $(SERVICE)),)
	$(error SERVICE is required, e.g. "make lint SERVICE=trainer". Valid services: $(SERVICES))
endif
ifeq ($(filter $(SERVICE),$(SERVICES)),)
	$(error unknown SERVICE "$(SERVICE)". Valid services: $(SERVICES))
endif
	cd internal/$(SERVICE) && golangci-lint run --timeout=5m
