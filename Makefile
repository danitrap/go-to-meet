BINARY_NAME := gotomeet
BUILD_DIR := build
GO_FILES := $(shell find . -name '*.go')

.PHONY: all build clean run

all: build

build: $(GO_FILES)
	@mkdir -p $(BUILD_DIR)
	go build -ldflags="-s -w" -o $(BUILD_DIR)/$(BINARY_NAME) cmd/gotomeet/main.go

clean:
	@rm -rf $(BUILD_DIR)

run:
	@go run cmd/gotomeet/main.go
