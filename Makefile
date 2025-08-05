# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GORUN=$(GOCMD) run
BINARY_NAME=xconnect
BINARY_UNIX=$(BINARY_NAME)_unix

# Build flags
LDFLAGS=-ldflags "-s -w"
BUILD_FLAGS=-a -installsuffix cgo

.PHONY: all build clean test coverage deps lint fmt vet security run debug

all: test build

build:
	CGO_ENABLED=0 GOOS=linux $(GOBUILD) $(BUILD_FLAGS) $(LDFLAGS) -o $(BINARY_NAME) ./cmd/main.go

build-local:
	$(GOBUILD) $(LDFLAGS) -o $(BINARY_NAME) ./cmd/main.go

run:
	$(GORUN) .

debug:
	dlv debug .

clean:
	$(GOCLEAN)
	rm -f $(BINARY_NAME)
	rm -f $(BINARY_UNIX)

test:
	$(GOTEST) -v ./...

coverage:
	$(GOTEST) -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out

deps:
	$(GOMOD) download
	$(GOMOD) tidy

lint:
	golangci-lint run

fmt:
	$(GOCMD) fmt ./...

vet:
	$(GOCMD) vet ./...

security:
	gosec ./...

docker-build:
	docker build -t $(BINARY_NAME):latest .

docker-run:
	docker run -p 8080:8080 $(BINARY_NAME):latest

install-tools:
	$(GOGET) -u github.com/golangci/golangci-lint/cmd/golangci-lint
	$(GOGET) -u github.com/securecodewarrior/gosec/v2/cmd/gosec
	$(GOGET) -u github.com/go-delve/delve/cmd/dlv

release: clean test lint vet security build

dev: deps fmt vet test build-local

help:
	@echo "Available targets:"
	@echo "  build        - Build production binary"
	@echo "  build-local  - Build local binary"
	@echo "  run          - Run application with 'go run .'"
	@echo "  debug        - Debug application with Delve"
	@echo "  clean        - Clean build artifacts"
	@echo "  test         - Run tests"
	@echo "  coverage     - Generate test coverage report"
	@echo "  deps         - Download and tidy dependencies"
	@echo "  lint         - Run linter"
	@echo "  fmt          - Format code"
	@echo "  vet          - Run go vet"
	@echo "  security     - Run security scan"
	@echo "  release      - Full release build with checks"
	@echo "  dev          - Development build"
	@echo "  install-tools- Install required tools (includes Delve)"