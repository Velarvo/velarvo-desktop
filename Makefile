.PHONY: dev build build-linux build-windows build-macos-arm64 build-macos-amd64 build-macos-universal \
        lint lint-go lint-frontend fmt fmt-go fmt-frontend format-check test test-go test-frontend \
        test-cover test-cover-html fe-build check

APP_NAME := Velarvo
FRONTEND_DIR := frontend
GO_CACHE := $(CURDIR)/.cache/go-build
GOLANGCI_LINT_CACHE := $(CURDIR)/.cache/golangci-lint
GO_TEST_ENV := GOCACHE=$(GO_CACHE)
GO_LINT_ENV := GOCACHE=$(GO_CACHE) GOLANGCI_LINT_CACHE=$(GOLANGCI_LINT_CACHE)
GO_PACKAGES := ./ ./internal/...

dev:
	wails dev

build:
	wails build -o $(APP_NAME) -clean

build-linux:
	PKG_CONFIG_PATH="/usr/lib/x86_64-linux-gnu/pkgconfig" \
	wails build -platform linux/amd64 -o $(APP_NAME) -clean -tags webkit2_41

build-windows:
	wails build -platform windows/amd64 -o $(APP_NAME) -clean

build-macos-arm64:
	wails build -platform darwin/arm64 -o $(APP_NAME) -clean

build-macos-amd64:
	wails build -platform darwin/amd64 -o $(APP_NAME) -clean

build-macos-universal:
	wails build -platform darwin/universal -o $(APP_NAME) -clean

lint: lint-go lint-frontend

lint-go:
	$(GO_LINT_ENV) golangci-lint run --config .golangci.yml $(GO_PACKAGES)

lint-frontend:
	cd $(FRONTEND_DIR) && bun run lint

fmt: fmt-go fmt-frontend

fmt-go:
	gofmt -w main.go trafficlight_*.go internal/

fmt-frontend:
	cd $(FRONTEND_DIR) && bun run format

format-check:
	cd $(FRONTEND_DIR) && bun run format:check

test: test-go test-frontend

test-go:
	$(GO_TEST_ENV) go test $(GO_PACKAGES) -race -v

test-frontend:
	cd $(FRONTEND_DIR) && bun run test

test-cover:
	$(GO_TEST_ENV) go test $(GO_PACKAGES) -race -coverprofile=coverage.out
	go tool cover -func=coverage.out

test-cover-html:
	$(GO_TEST_ENV) go test $(GO_PACKAGES) -race -coverprofile=coverage.out
	go tool cover -html=coverage.out

fe-build:
	cd $(FRONTEND_DIR) && bun run build

check: format-check lint test fe-build
