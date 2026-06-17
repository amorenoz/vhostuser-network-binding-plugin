PLUGIN_NAME := vhostuser-network-binding-plugin
BIN_DIR ?= bin
BIN     ?= $(BIN_DIR)/$(PLUGIN_NAME)

GO      ?= go
LINT    := golangci-lint

GOLANG_VERSION  ?= 1.26

CONTAINER_TOOL 	?= podman
REGISTRY       	?= quay.io/amorenoz
IMAGE_NAME     	?= $(REGISTRY)/$(PLUGIN_NAME)
IMAGE_TAG      	?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")

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

build-image:
	$(CONTAINER_TOOL) build \
		--build-arg GOLANG_VERSION=$(GOLANG_VERSION) \
		-t $(IMAGE_NAME):$(IMAGE_TAG) \
		-f $(CURDIR)/Dockerfile \
		$(CURDIR)

push-image: build-image
	$(CONTAINER_TOOL) push $(IMAGE_NAME):$(IMAGE_TAG)
