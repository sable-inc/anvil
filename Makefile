.PHONY: build test lint generate install clean check

BINARY  := anvil
GO      := go1.26.0
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT  := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
DATE    := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS := -ldflags "-s -w \
	-X github.com/sable-inc/anvil/internal/version.Version=$(VERSION) \
	-X github.com/sable-inc/anvil/internal/version.Commit=$(COMMIT) \
	-X github.com/sable-inc/anvil/internal/version.Date=$(DATE)"

build:
	$(GO) build $(LDFLAGS) -o bin/$(BINARY) ./cmd/anvil

test:
	$(GO) test -race -count=1 ./...

lint:
	$(shell go env GOPATH)/bin/golangci-lint run

generate:
	@echo "Fetching OpenAPI spec from sable-api..."
	@mkdir -p api
	curl -sf http://localhost:8080/openapi.json -o api/openapi.json
	oapi-codegen -generate types,client -package api api/openapi.json > api/generated.go
	@echo "Generated api/generated.go"

install:
	$(GO) install $(LDFLAGS) ./cmd/anvil

clean:
	rm -rf bin/ api/openapi.json

check: build test lint
	@echo "All checks passed."
