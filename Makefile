.PHONY: build install dev lint clean tidy

BINARY_NAME = ca
BUILD_DIR   = bin

build:
	@mkdir -p $(BUILD_DIR)
	go build -ldflags="-s -w" -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/ca/
	@echo "✅ Built: $(BUILD_DIR)/$(BINARY_NAME)"

install:
	go install ./cmd/ca/
	@echo "✅ Installed to $(shell go env GOPATH)/bin/$(BINARY_NAME)"

dev:
	go run ./cmd/ca/ .

lint:
	golangci-lint run ./...

clean:
	rm -rf $(BUILD_DIR)

tidy:
	go mod tidy
