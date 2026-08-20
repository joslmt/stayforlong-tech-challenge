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

.PHONY: help banner start stop logs build run fmt fmt-check vet lint vuln test tests test-unit test-integration e2e

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

banner:
ifeq ($(OS),Windows_NT)
	${info   ____ _____  _ __   _______ ___  ____  _     ___  _   _  ____}
	${info  / ___|_   _|/ \ \ / /  ___/ _ \|  _ \| |   / _ \| \ | |/ ___|}
	${info  \___ \ | | / _ \ V /| |_ | | | | |_) | |  | | | |  \| | |  _}
	${info   ___) || |/ ___ \| | |  _|| |_| |  _ <| |__| |_| | |\  | |_| |}
	${info  |____/ |_/_/   \_\_| |_|   \___/|_| \_\_____\___/|_| \_|\____|}
	${info }
	${info    ____ _   _    _    _     _     _____ _   _  ____ _____}
	${info   / ___| | | |  / \  | |   | |   | ____| \ | |/ ___| ____|}
	${info  | |   | |_| | / _ \ | |   | |   |  _| |  \| | |  _|  _|}
	${info  | |___|  _  |/ ___ \| |___| |___| |___| |\  | |_| | |___}
	${info   \____|_| |_/_/   \_\_____|_____|_____|_| \_|\____|_____|}
	@echo Playground: http://localhost:$(PORT)
else
	@printf '\033[38;2;237;36;113m%s\n' '  ____ _____  _ __   _______ ___  ____  _     ___  _   _  ____'
	@printf '%s\n' ' / ___|_   _|/ \ \ / /  ___/ _ \|  _ \| |   / _ \| \ | |/ ___|'
	@printf '%s\n' ' \___ \ | | / _ \ V /| |_ | | | | |_) | |  | | | |  \| | |  _'
	@printf '%s\n' '  ___) || |/ ___ \| | |  _|| |_| |  _ <| |__| |_| | |\  | |_| |'
	@printf '%s\n' ' |____/ |_/_/   \_\_| |_|   \___/|_| \_\_____\___/|_| \_|\____|'
	@printf '\n'
	@printf '%s\n' '   ____ _   _    _    _     _     _____ _   _  ____ _____'
	@printf '%s\n' '  / ___| | | |  / \  | |   | |   | ____| \ | |/ ___| ____|'
	@printf '%s\n' ' | |   | |_| | / _ \ | |   | |   |  _| |  \| | |  _|  _|'
	@printf '%s\n' ' | |___|  _  |/ ___ \| |___| |___| |___| |\  | |_| | |___'
	@printf '%s\033[0m\n' '  \____|_| |_/_/   \_\_____|_____|_____|_| \_|\____|_____|'
	@printf 'Playground: http://localhost:$(PORT)\n'
endif

start: banner
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
