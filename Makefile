.PHONY: all build test vet lint clean install

GO := go
BINARY := proto-migrate
CMD_PATH := ./cmd/$(BINARY)

# Version information
VERSION ?= dev
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
DATE := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

# Build flags
LDFLAGS := -X 'github.com/jackchuka/proto-migrate/internal/version.Version=$(VERSION)'
LDFLAGS += -X 'github.com/jackchuka/proto-migrate/internal/version.Commit=$(COMMIT)'
LDFLAGS += -X 'github.com/jackchuka/proto-migrate/internal/version.Date=$(DATE)'

all: build

build:
	$(GO) build -ldflags "$(LDFLAGS)" -o $(BINARY) $(CMD_PATH)

install:
	$(GO) install -ldflags "$(LDFLAGS)" $(CMD_PATH)

test:
	$(GO) test ./... -v -cover

vet:
	$(GO) vet ./...

lint:
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not installed, skipping"; \
	fi

clean:
	rm -f $(BINARY)
	$(GO) clean -cache

fmt:
	$(GO) fmt ./...

tidy:
	$(GO) mod tidy

ci: fmt tidy vet test build

help:
	@echo "Available targets:"
	@echo "  build     - Build the binary"
	@echo "  install   - Install the binary"
	@echo "  test      - Run tests"
	@echo "  vet       - Run go vet"
	@echo "  lint      - Run golangci-lint (if installed)"
	@echo "  clean     - Clean build artifacts"
	@echo "  fmt       - Format code"
	@echo "  tidy      - Tidy go modules"
	@echo "  ci        - Run CI checks (fmt, tidy, vet, test, build)"
