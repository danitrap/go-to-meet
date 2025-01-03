BINARY_NAME := go-to-meet
BUILD_DIR := build
GO_FILES := $(shell find . -name '*.go')

.PHONY: all build clean run delete-token

all: build

build: $(GO_FILES)
	@mkdir -p $(BUILD_DIR)
	go build -ldflags="-s -w" -o $(BUILD_DIR)/$(BINARY_NAME) cmd/go-to-meet/main.go

clean:
	@rm -rf $(BUILD_DIR)

run:
	@go run cmd/go-to-meet/main.go

delete-token:
	@rm -f $(HOME)/Library/Application\ Support/go-to-meet/token.json
