include .bingo/Variables.mk

OS                    ?= $(shell uname -s | tr '[A-Z]' '[a-z]')
ARCH                  ?= $(shell uname -m)

VERSION               := $(strip $(shell [ -d .git ] && git describe --abbrev=0))
LONG_VERSION          := $(strip $(shell [ -d .git ] && git describe --always --tags --dirty))
BUILD_DATE            := $(shell date -u +"%Y-%m-%dT%H:%M:%S%Z")
VCS_REF               := $(strip $(shell [ -d .git ] && git rev-parse HEAD))

GO_PACKAGES            = $(shell go list ./... | grep -v -E '/vendor/|/test')
GO_FILES              := $(shell find . -name \*.go -print)

GOBUILD               := GOOS=$(OS) GOARCH=$(ARCH) go build -mod=vendor
GOINSTALL             := go install -mod=vendor
GOMOD                 := go mod
GOFMT                 := gofmt
LDFLAGS               := '-s -w -X main.version=$(VERSION) -X main.commit=$(VCS_REF) -X main.date=$(BUILD_DATE)'
TAGS                  := netgo

ROOT_DIR              := $(CURDIR)
BIN_DIR               ?= $(ROOT_DIR)/tmp/bin
UPX                   := upx

DOCKER                := docker
DOCKER_BUILD          := $(DOCKER) build
DOCKER_PUSH           := $(DOCKER) push
DOCKER_COMPOSE        := docker-compose

CONTAINER_REPO        ?=  meltwater/drone-cache

V                      = 0
Q                      = $(if $(filter 1,$V),,@)
M                      = $(shell printf "\033[34;1m▶\033[0m")

.PHONY: default all
default: drone-cache
all: drone-cache

.PHONY: setup
setup: ## Setups dev environment
setup: vendor ; $(info $(M) running setup for development )
	$(Q) make $(GOTEST) $(EMBEDMD) $(LICHE) $(GOLANGCI_LINT) $(BINGO)

drone-cache: ## Runs drone-cache target
drone-cache: vendor main.go $(wildcard *.go) $(wildcard */*.go) ; $(info $(M) running drone-cache )
	$(Q) CGO_ENABLED=0 $(GOBUILD) -a -ldflags $(LDFLAGS) -tags $(TAGS) -o $@ .

.PHONY: clean
clean: ## Cleans build resourcess
clean: ; $(info $(M) running clean )
	$(Q) rm -f drone-cache
	$(Q) rm -rf target
	$(Q) rm -rf tmp

tmp/help.txt: drone-cache
	-mkdir -p tmp
	$(ROOT_DIR)/drone-cache --help &> tmp/help.txt

tmp/make_help.txt: Makefile
	-mkdir -p tmp
	$(Q) awk 'BEGIN {FS = ":.*##"; printf "Usage:\n  make <target>\n\nTargets:\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  %-15s\t %s\n", $$1, $$2 }' $(MAKEFILE_LIST) &> tmp/make_help.txt

README.md: tmp/help.txt tmp/make_help.txt $(EMBEDMD)
	$(EMBEDMD) -w README.md

tmp/docs.txt: drone-cache
	$(Q) echo "IMPLEMENT ME"

DOCS.md: tmp/docs.txt $(EMBEDMD)
	$(EMBEDMD) -w DOCS.md

docs: ## Generates docs
docs: clean README.md DOCS.md $(LICHE)
	$(Q) $(LICHE) --recursive docs --document-root .
	$(Q) $(LICHE) --exclude "(goreportcard.com)" --document-root . *.md

generate: ## Generate documentation, website and yaml files,
generate: docs # site
	$(Q) echo "Generated!"

.PHONY: vendor
vendor: ## Updates vendored copy of dependencies
vendor: ; $(info $(M) running vendor )
	$(Q) $(GOMOD) tidy
	$(Q) $(GOMOD) vendor -v

.PHONY: compress
compress: ## Creates compressed binary
compress: drone-cache ; $(info $(M) running compress )
	# Add as dependency
	$(Q) $(UPX) drone-cache

.PHONY: container
container: ## Builds drone-cache docker image with latest tag
container: drone-cache Dockerfile ; $(info $(M) running container )
	$(Q) $(DOCKER_BUILD) -t $(CONTAINER_REPO):$(LONG_VERSION) .

.PHONY: container-push
container-push: ## Pushes latest $(CONTAINER_REPO) image to repository
container-push: container ; $(info $(M) running container-push )
	$(Q) $(DOCKER_PUSH) $(CONTAINER_REPO):$(LONG_VERSION)

.PHONY: test
test: ## Runs tests
test: $(GOTEST) ; $(info $(M) running test)
	$(DOCKER_COMPOSE) up -d && sleep 1
	-$(GOTEST) -race -short -cover -failfast -tags=integration ./...
	$(DOCKER_COMPOSE) down -v

.PHONY: test-integration ; $(info $(M) running test-integration )
test-integration: ## Runs integration tests
test-integration: $(GOTEST)
	$(DOCKER_COMPOSE) up -d && sleep 1
	-$(GOTEST) -race -cover -tags=integration -v ./...
	$(DOCKER_COMPOSE) down -v

.PHONY: test-unit
test-unit: ## Runs unit tests
test-unit: $(GOTEST) ; $(info $(M) running test-unit )
	$(GOTEST) -race -cover -benchmem -v ./...

.PHONY: test-e2e
test-e2e: ## Runs e2e tests
test-e2e: $(GOTEST) ; $(info $(M) running test-e2e )
	$(DOCKER_COMPOSE) up -d && sleep 1
	-$(GOTEST) -race -cover -tags=integration -v ./internal/plugin
	$(DOCKER_COMPOSE) down -v

.PHONY: lint
lint: ## Runs golangci-lint analysis
lint: $(GOLANGCI_LINT) ; $(info $(M) running lint )
	# Check .golangci.yml for configuration
	$(Q) $(GOLANGCI_LINT) run -v --enable-all --skip-dirs tmp -c .golangci.yml

.PHONY: fix
fix: ## Runs golangci-lint fix
fix: $(GOLANGCI_LINT) format ; $(info $(M) running fix )
	$(Q) $(GOLANGCI_LINT) run --fix --enable-all --skip-dirs tmp -c .golangci.yml

.PHONY: format
format: ## Runs gofmt
format: ; $(info $(M) running format )
	$(Q) $(GOFMT) -w -s $(GO_FILES)

.PHONY: help
help: ## Shows this help message
	$(Q) awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n\nTargets:\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-15s\033[0m\t %s\n", $$1, $$2 }' $(MAKEFILE_LIST)
