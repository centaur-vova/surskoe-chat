.PHONY: lint
lint:
	golangci-lint run ./...

.PHONY: lint-fix
lint-fix:
	golangci-lint run --fix ./...

.PHONY: fmt
fmt:
	go fmt ./...

.PHONY: tidy
tidy:
	go mod tidy

.PHONY: test
test:
	go test -v -race ./...

.PHONY: build
build:
	go build -o bin/bot cmd/bot/main.go

.PHONY: run
run:
	go run cmd/bot/main.go

.PHONY: all
all: tidy fmt lint test build
