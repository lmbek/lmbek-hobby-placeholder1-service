# User Service Makefile

.PHONY: test build run

test:
	@echo "==> Running tests for user-service..."
	go test -v ./...

build:
	@echo "==> Building user-service..."
	go build -o user-service main.go

run:
	@echo "==> Running user-service..."
	go run main.go
