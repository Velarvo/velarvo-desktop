.PHONY: install dev build build-linux build-windows build-macos-arm64 build-macos-amd64 build-macos-universal \
        lint lint-go lint-frontend fmt fmt-go fmt-frontend fmt-generated \
        format-check format-check-go format-check-frontend format-check-generated \
        test test-go test-frontend test-cover test-cover-html fe-build \
        typecheck-frontend verify-go-mod audit-frontend precommit prepush check ci

APP_NAME := Velarvo
FRONTEND_DIR := frontend
GO_CACHE := $(CURDIR)/.cache/go-build
GOLANGCI_LINT_CACHE := $(CURDIR)/.cache/golangci-lint
GO_TEST_ENV := GOCACHE=$(GO_CACHE)
GO_LINT_ENV := GOCACHE=$(GO_CACHE) GOLANGCI_LINT_CACHE=$(GOLANGCI_LINT_CACHE)
GO_LINT_PACKAGES := ./ ./internal/...
GO_TEST_PACKAGES := ./internal/...
STAGED_GO_FILES := $(shell git diff --cached --name-only --diff-filter=ACMR -- '*.go')

install:
	bun install --frozen-lockfile
	cd $(FRONTEND_DIR) && bun install --frozen-lockfile

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
	$(GO_LINT_ENV) golangci-lint run --config .golangci.yml $(GO_LINT_PACKAGES)

lint-frontend:
	cd $(FRONTEND_DIR) && bun run lint

typecheck-frontend:
	cd $(FRONTEND_DIR) && bun run typecheck

fmt: fmt-go fmt-frontend fmt-generated

fmt-go:
	$(GO_LINT_ENV) golangci-lint fmt --config .golangci.yml

fmt-frontend:
	cd $(FRONTEND_DIR) && bun run format

fmt-generated:
	@find $(FRONTEND_DIR)/wailsjs -type f \( -name '*.ts' -o -name '*.js' \) -print0 | xargs -0 perl -0pi -e 's/[ \t]+\r?\n/\n/g'

format-check: format-check-go format-check-frontend format-check-generated

format-check-go:
	$(GO_LINT_ENV) golangci-lint fmt --config .golangci.yml --diff

format-check-frontend:
	cd $(FRONTEND_DIR) && bun run format:check

format-check-generated:
	@find $(FRONTEND_DIR)/wailsjs -type f \( -name '*.ts' -o -name '*.js' \) -print0 | xargs -0 perl -ne 'if (/[ \t]+$$/) { print "$$ARGV:$$.: trailing whitespace\n"; $$bad=1 } END { exit($$bad ? 1 : 0) }'

test: test-go test-frontend

test-go:
	$(GO_TEST_ENV) go test $(GO_TEST_PACKAGES) -race -v

test-frontend:
	cd $(FRONTEND_DIR) && bun run test

test-cover:
	$(GO_TEST_ENV) go test $(GO_TEST_PACKAGES) -race -coverprofile=coverage.out
	go tool cover -func=coverage.out

test-cover-html:
	$(GO_TEST_ENV) go test $(GO_TEST_PACKAGES) -race -coverprofile=coverage.out
	go tool cover -html=coverage.out

fe-build:
	cd $(FRONTEND_DIR) && bun run build

verify-go-mod:
	go mod verify

audit-frontend:
	cd $(FRONTEND_DIR) && bun run audit

precommit:
	cd $(FRONTEND_DIR) && bun run --silent lint:staged
	@if [ -n "$(STAGED_GO_FILES)" ]; then \
		for file in $(STAGED_GO_FILES); do \
			tmp=$$(mktemp); \
			golangci-lint fmt --config .golangci.yml --stdin < "$$file" > "$$tmp"; \
			cat "$$tmp" > "$$file"; \
			rm -f "$$tmp"; \
			git add -- "$$file"; \
		done; \
	fi
	$(MAKE) format-check-go format-check-generated lint-go

prepush:
	$(MAKE) test fe-build

check: format-check lint typecheck-frontend verify-go-mod test fe-build
ci: check
