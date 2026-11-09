.PHONY: help run build test cover test-integration lint tidy \
	docker-build docker-build-deps docker-build-builder docker-run up down

APP_NAME ?= web-analyzer
IMAGE ?= web-analyzer
BIN ?= bin/server
MAIN_PKG ?= ./cmd/server
GO_VERSION ?= 1.27
COVER_MIN ?= 70
COVER_PROFILE ?= tmp/cover.out
TEST_PKGS ?= ./internal/... ./pkg/...
INTEGRATION_PKGS ?= $(TEST_PKGS) ./tests/...

## list targets
help: 
	@grep -E '^[a-zA-Z0-9_.-]+:.*## ' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*## "}; {printf "  \033[36m%-22s\033[0m %s\n", $$1, $$2}'

## run the HTTP server locally
run: 
	go run $(MAIN_PKG)

## compile server binary to $(BIN)
build: 
	@mkdir -p $(dir $(BIN))
	CGO_ENABLED=0 go build -o $(BIN) $(MAIN_PKG)

## run unit tests with race detector
test: 
	go test -race $(TEST_PKGS)

## unit test coverage; fail if total under $(COVER_MIN)%
cover: 
	@mkdir -p $(dir $(COVER_PROFILE))
	go test -race -covermode=atomic -coverprofile=$(COVER_PROFILE) $(TEST_PKGS)
	@total=$$(go tool cover -func=$(COVER_PROFILE) | awk '/total:/ {gsub(/%/,""); print $$3}'); \
		echo "total coverage: $${total}% (min $(COVER_MIN)%)"; \
		awk -v t="$$total" -v min="$(COVER_MIN)" 'BEGIN {exit !(t+0 >= min+0)}' || \
		(echo "coverage below minimum" && exit 1)

## integration tests (go test -tags=integration)
test-integration: 
	go test -race -tags=integration $(INTEGRATION_PKGS)

## go vet and golangci-lint when installed
lint: 
	go vet ./...
	@command -v golangci-lint >/dev/null 2>&1 && golangci-lint run ./... || \
		echo "golangci-lint not installed; skipped"

## go mod tidy
tidy: 
	go mod tidy

## build Docker deps stage (module download cache)
docker-build-deps: 
	docker build --build-arg GO_VERSION=$(GO_VERSION) --target deps -t $(IMAGE):deps .

## build Docker builder stage (compile only)
docker-build-builder: 
	docker build --build-arg GO_VERSION=$(GO_VERSION) --target builder -t $(IMAGE):builder .

## build production runtime image
docker-build: 
	docker build --build-arg GO_VERSION=$(GO_VERSION) -t $(IMAGE):latest .

## run container on host :8080
docker-run: docker-build 
	docker run --rm -p 8080:8080 -e PORT=8080 $(IMAGE):latest

## docker compose up (build + detached)
up: 
	docker compose up --build -d

## docker compose down
down: 
	docker compose down