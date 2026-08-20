.DEFAULT_GOAL := help

PORT ?= 8080
HTTP_ADDRESS ?= :$(PORT)
export PORT
export HTTP_ADDRESS

PRODUCTION_PACKAGES := ./cmd/api,./internal/booking/domain,./internal/booking/application,./internal/booking/adapters/in/http,./internal/bootstrap,./internal/platform/httpserver,./web
TEST_PACKAGES := ./cmd/api ./internal/booking/domain ./internal/booking/application ./internal/booking/adapters/in/http ./internal/bootstrap ./internal/platform/httpserver

ifeq ($(OS),Windows_NT)
	BINARY := bin/stayforlong.exe
	CREATE_BIN := if not exist bin mkdir bin
else
	BINARY := bin/stayforlong
	CREATE_BIN := mkdir -p bin
endif

.PHONY: help start stop logs build run fmt fmt-check vet lint vuln test tests test-unit test-integration e2e

help:
	@echo Usage: make target
	@echo Available targets:
	@echo   start              Build and start the complete project
	@echo   stop               Stop and remove local containers
	@echo   logs               Follow application logs
	@echo   build              Build the Go binary
	@echo   run                Run the API locally
	@echo   fmt                Format Go source files
	@echo   fmt-check          Verify Go source formatting
	@echo   vet                Run Go static analysis
	@echo   lint               Run golangci-lint
	@echo   vuln               Scan Go dependencies for vulnerabilities
	@echo   test               Run tests and report production coverage
	@echo   tests              Alias for test
	@echo   test-unit          Run domain and application unit tests
	@echo   test-integration   Run HTTP integration tests
	@echo   e2e                Run production-container E2E tests

start:
	docker compose up --build

stop:
	docker compose down --remove-orphans

logs:
	docker compose logs --follow api

build:
	$(CREATE_BIN)
	go build -mod=readonly -trimpath -o $(BINARY) ./cmd/api

run:
	go run -mod=readonly ./cmd/api

fmt:
	go run ./tools/gofmtcheck -write

fmt-check:
	go run ./tools/gofmtcheck

vet:
	go vet ./...

lint:
	golangci-lint run ./...

vuln:
	govulncheck ./...

test:
	go test -mod=readonly -race -covermode=atomic -coverpkg=$(PRODUCTION_PACKAGES) -coverprofile=coverage.out $(TEST_PACKAGES)
	@echo Aggregate production coverage:
	go tool cover -func=coverage.out

tests: test

test-unit:
	go test -mod=readonly -race ./internal/booking/domain ./internal/booking/application

test-integration:
	go test -mod=readonly -race -coverpkg=./internal/booking/adapters/in/http,./internal/bootstrap,./internal/platform/httpserver ./internal/booking/adapters/in/http ./internal/bootstrap ./internal/platform/httpserver

e2e:
	go test -mod=readonly -tags=e2e -count=1 -v ./test/e2e
