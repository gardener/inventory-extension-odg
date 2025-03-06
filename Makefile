.DEFAULT_GOAL                := build
REPO_ROOT                    := $(shell go list -m -f '{{ .Dir }}')
LOCAL_BIN                    ?= $(REPO_ROOT)/bin
BINARY                       ?= $(LOCAL_BIN)/inventory-extension-odg
SRC_DIRS                     := $(shell go list -f '{{ .Dir }}' ./...)
VERSION                      := $(shell cat VERSION)

EFFECTIVE_VERSION            ?= $(VERSION)-$(shell git rev-parse --short HEAD)
ifneq ($(strip $(shell git status --porcelain 2>/dev/null)),)
	EFFECTIVE_VERSION := $(EFFECTIVE_VERSION)-dirty
endif

IMAGE                        ?= europe-docker.pkg.dev/gardener-project/releases/gardener/inventory-extension-odg
IMAGE_TAG                    ?= $(EFFECTIVE_VERSION)

$(LOCAL_BIN):
	mkdir -p $(LOCAL_BIN)

.PHONY: goimports-reviser
goimports-reviser:
	go tool goimports-reviser -set-exit-status -rm-unused ./...

.PHONY: lint
lint:
	go tool golangci-lint run --config=$(REPO_ROOT)/.golangci.yaml ./...

$(BINARY): $(SRC_DIRS) | $(LOCAL_BIN)
	go build \
		-o $(BINARY) \
		-ldflags="-X 'github.tools.sap/kubernetes/inventory-extension-odg/pkg/version.Version=${EFFECTIVE_VERSION}'" \
		./cmd/inventory-extension-odg

.PHONY: build
build: $(BINARY)

.PHONY: get
get:
	go mod download
	go mod tidy

.PHONY: test
test:
	go test -v -race ./...

.PHONY: test-cover
test-cover:
	go test -v -race -coverprofile=coverage.txt -covermode=atomic ./...

.PHONY: docker-build
docker-build:
	docker build -t $(IMAGE):$(IMAGE_TAG) -t $(IMAGE):latest .

.PHONY: docker-compose-up
docker-compose-up:
	docker compose up --build --remove-orphans
