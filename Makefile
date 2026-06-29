PLUGIN_NAME := vhostuser-network-binding-plugin
BIN_DIR ?= $(CURDIR)/bin
BUILD_DIR ?= $(CURDIR)/build

OUT ?= $(BUILD_DIR)/$(PLUGIN_NAME)

GO      ?= go
LINT    := $(BIN_DIR)/golangci-lint

GOLANG_VERSION  ?= 1.26
GOLANGCI_LINT_VERSION ?= v2.7.2

CONTAINER_TOOL 	?= podman
REGISTRY       	?= quay.io/kubevirt
IMAGE_NAME     	?= $(REGISTRY)/$(PLUGIN_NAME)
IMAGE_TAG      	?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")

.PHONY: all build test lint clean

all: build

build:
	$(GO) build -o $(OUT) ./cmd/...

test:
	$(GO) test ./pkg/...

lint: $(LINT)
	$(LINT) run ./...

$(LINT):
	GOBIN=$(BIN_DIR) go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION)

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
