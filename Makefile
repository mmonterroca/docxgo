SHELL := /bin/bash

BINARY_NAME ?= docxgo
BUILD_DIR ?= bin
COVERAGE_PROFILE ?= coverage.out
COVERAGE_HTML ?= coverage.html
DOTNET_ARTIFACTS ?= /tmp/docxgo-dotnet-artifacts

.PHONY: help test build install run coverage dotnet-build examples check clean

help: ## Show available targets
	@awk 'BEGIN {FS = ":.*## "; printf "Available targets:\n"} /^[a-zA-Z0-9_-]+:.*## / {printf "  %-14s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

test: ## Run all Go tests
	go test ./...

build: ## Build the docxgo CLI into bin/
	mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/docxgo

install: ## Install the docxgo CLI into GOPATH/bin
	go install ./cmd/docxgo

run: ## Run the docxgo CLI from source, for example: make run ARGS="version"
	go run ./cmd/docxgo $(ARGS)

coverage: ## Generate Go coverage profile and HTML report
	go test -coverprofile=$(COVERAGE_PROFILE) ./...
	go tool cover -html=$(COVERAGE_PROFILE) -o $(COVERAGE_HTML)

dotnet-build: ## Build the Open XML validator
	dotnet build DocxValidator/DocxValidator.csproj --artifacts-path $(DOTNET_ARTIFACTS)

examples: ## Run example programs and optional Open XML validation
	./examples/run_all_examples.sh

check: test build dotnet-build ## Run the core validation suite

clean: ## Remove local build and coverage artifacts
	rm -rf $(BUILD_DIR) $(COVERAGE_PROFILE) $(COVERAGE_HTML)
