BIN_DIR ?= bin
BIN     ?= $(BIN_DIR)/vhostuser-network-binding-plugin

GO      ?= go
LINT    := golangci-lint

.PHONY: all build test lint clean

all: build

build:
	$(GO) build -o $(BIN) ./cmd/...

test:
	$(GO) test ./pkg/...

lint:
	$(LINT) run ./...

clean:
	rm -rf $(BIN_DIR)
