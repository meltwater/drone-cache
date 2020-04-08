ROOT_DIR              := $(CURDIR)
SCRIPTS               := $(ROOT_DIR)/scripts

VERSION               := $(strip $(shell [ -d .git ] && git describe --always --tags --dirty))
BUILD_DATE            := $(shell date -u +"%Y-%m-%dT%H:%M:%S%Z")
VCS_REF               := $(strip $(shell [ -d .git ] && git rev-parse --short HEAD))

GO_PACKAGES            = $(shell go list ./... | grep -v -E '/vendor/|/test')
GO_FILES              := $(shell find . -name \*.go -print)

GOPATH                := $(firstword $(subst :, ,$(shell go env GOPATH)))
GOBIN                 := $(GOPATH)/bin

GOCMD                 := go
GOBUILD               := $(GOCMD) build
GOMOD                 := $(GOCMD) mod
GOGET                 := $(GOCMD) get
GOFMT                 := gofmt

GOLANGCI_LINT_VERSION  = v1.21.0
GOLANGCI_LINT_BIN      = $(GOBIN)/golangci-lint
EMBEDMD_BIN            = $(GOBIN)/embedmd
GOTEST_BIN             = $(GOBIN)/gotest
GORELEASER_VERSION     = v0.131.1
GORELEASER_BIN         = $(GOBIN)/goreleaser
LICHE_BIN              = $(GOBIN)/liche

UPX                   := upx

DOCKER                := docker
DOCKER_BUILD          := $(DOCKER) build
DOCKER_PUSH           := $(DOCKER) push
DOCKER_COMPOSE        := docker-compose

V                      = 0
Q                      = $(if $(filter 1,$V),,@)
M                      = $(shell printf "\033[34;1m▶\033[0m")

.PHONY: default all
default: drone-cache
all: drone-cache

.PHONY: setup
setup: ## Setups dev environment
setup: ; $(info $(M) running setup )
	$(Q) $(SCRIPTS)/setup_dev_environment.sh

drone-cache: ## Runs drone-cache target
drone-cache: vendor main.go $(wildcard *.go) $(wildcard */*.go) ; $(info $(M) running drone-cache )
	$(Q) CGO_ENABLED=0 $(GOBUILD) -mod=vendor -a -tags netgo -ldflags '-s -w -X main.version=$(VERSION)' -o $@ .

.PHONY: build
build: ## Runs build target
build: main.go $(wildcard *.go) $(wildcard */*.go) ; $(info $(M) running build )
	$(Q) $(GOBUILD) -mod=vendor -tags netgo -ldflags '-X main.version=$(VERSION)' -o drone-cache .

.PHONY: release
release: ## Release dron-cache
release: drone-cache $(GORELEASER_BIN) ; $(info $(M) running release )
	$(Q) $(GORELEASER_BIN) release --rm-dist

.PHONY: snapshot
snapshot: ## Creates snapshot release without publishing it
snapshot: drone-cache $(GORELEASER_BIN); $(info $(M) running snapshot )
	$(Q) $(GORELEASER_BIN) release --skip-publish --rm-dist --snapshot

.PHONY: clean
clean: ## Cleans build resourcess
clean: ; $(info $(M) running clean )
	$(Q) rm -f drone-cache
	$(Q) rm -rf target

tmp/help.txt: drone-cache
	mkdir -p tmp
	$(ROOT_DIR)/drone-cache --help &> tmp/help.txt

README.md: tmp/help.txt
	$(EMBEDMD_BIN) -w README.md

tmp/docs.txt: drone-cache
	$(Q) echo "IMPLEMENT ME"

DOCS.md: tmp/docs.txt
	$(EMBEDMD_BIN) -w DOCS.md

docs: ## Generates docs
docs: clean README.md DOCS.md $(LICHE_BIN)
	$(Q) $(LICHE_BIN) --recursive docs --document-root .
	$(Q) $(LICHE_BIN) --exclude "(goreportcard.com)" --document-root . *.md

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
container: release Dockerfile ; $(info $(M) running container )
	$(Q) $(DOCKER_BUILD) --build-arg BUILD_DATE="$(BUILD_DATE)" \
		--build-arg VERSION="$(VERSION)" \
		--build-arg VCS_REF="$(VCS_REF)" \
		--build-arg DOCKERFILE_PATH="/Dockerfile" \
		-t meltwater/drone-cache:latest .

.PHONY: container-dev
container-dev: ## Builds development drone-cache docker image
container-dev: snapshot Dockerfile ; $(info $(M) running container-dev )
	$(Q) $(DOCKER_BUILD) --build-arg BUILD_DATE="$(BUILD_DATE)" \
		--build-arg VERSION="$(VERSION)" \
		--build-arg VCS_REF="$(VCS_REF)" \
		--build-arg DOCKERFILE_PATH="/Dockerfile" \
		--no-cache \
		-t meltwater/drone-cache:dev .

.PHONY: container-push
container-push: ## Pushes latest meltwater/drone-cache image to repository
container-push: container ; $(info $(M) running container-push )
	$(Q) $(DOCKER_PUSH) meltwater/drone-cache:latest

.PHONY: container-push-dev
container-push-dev: ## Pushes dev meltwater/drone-cache image to repository
container-push-dev: container-dev ; $(info $(M) running container-push-dev )
	$(Q) $(DOCKER_PUSH) meltwater/drone-cache:dev

.PHONY: test
test: ## Runs tests
test: $(GOTEST_BIN) ; $(info $(M) running test)
	$(DOCKER_COMPOSE) up -d && sleep 1
	-$(GOTEST_BIN) -race -short -cover -failfast -tags=integration ./...
	$(DOCKER_COMPOSE) down -v

.PHONY: test-integration ; $(info $(M) running test-integration )
test-integration: ## Runs integration tests
test-integration: $(GOTEST_BIN)
	$(DOCKER_COMPOSE) up -d && sleep 1
	-$(GOTEST_BIN) -race -cover -tags=integration -v ./...
	$(DOCKER_COMPOSE) down -v

.PHONY: test-unit
test-unit: ## Runs unit tests
test-unit: $(GOTEST_BIN) ; $(info $(M) running test-unit )
	$(GOTEST_BIN) -race -cover -benchmem -v ./...

.PHONY: test-e2e
test-e2e: ## Runs e2e tests
test-e2e: $(GOTEST_BIN) ; $(info $(M) running test-e2e )
	$(DOCKER_COMPOSE) up -d && sleep 1
	-$(GOTEST_BIN) -race -cover -tags=integration -v ./internal/plugin
	$(DOCKER_COMPOSE) down -v

.PHONY: lint
lint: ## Runs golangci-lint analysis
lint: $(GOLANGCI_LINT_BIN) ; $(info $(M) running lint )
	# Check .golangci.yml for configuration
	$(Q) $(GOLANGCI_LINT_BIN) run -v --enable-all --skip-dirs tmp -c .golangci.yml

.PHONY: fix
fix: ## Runs golangci-lint fix
fix: $(GOLANGCI_LINT_BIN) format ; $(info $(M) running fix )
	$(Q) $(GOLANGCI_LINT_BIN) run --fix --enable-all --skip-dirs tmp -c .golangci.yml

.PHONY: format
format: ## Runs gofmt
format: ; $(info $(M) running format )
	$(Q) $(GOFMT) -w -s $(GO_FILES)

.PHONY: help
help: ## Shows this help message
	$(Q) echo 'usage: make [target] ...'
	$(Q) echo
	$(Q) echo 'targets : '
	$(Q) echo
	$(Q) fgrep -h "##" $(MAKEFILE_LIST) | fgrep -v fgrep | sed -e 's/\\$$//' | sed -e 's/##//'| column -s: -t

$(GOTEST_BIN): ; $(info $(M) getting gotest )
	$(Q) GO111MODULE=off $(GOGET) -u github.com/rakyll/gotest

$(EMBEDMD_BIN): ; $(info $(M) getting embedmd )
	$(Q) GO111MODULE=off $(GOGET) -u github.com/campoy/embedmd

$(GOLANGCI_LINT_BIN): ; $(info $(M) getting golangci-lint )
	$(Q) curl -sfL https://raw.githubusercontent.com/golangci/golangci-lint/$(GOLANGCI_LINT_VERSION)/install.sh \
		| sed -e '/install -d/d' \
		| sh -s -- -b $(GOBIN) $(GOLANGCI_LINT_VERSION)

$(GORELEASER_BIN): ; $(info $(M) getting goreleaser )
	$(Q) curl -sfL https://install.goreleaser.com/github.com/goreleaser/goreleaser.sh \
		| VERSION=$(GORELEASER_VERSION) sh -s -- -b $(GOBIN) $(GORELEASER_VERSION)

$(LICHE_BIN): ; $(info $(M) getting liche )
	$(Q) GO111MODULE=on $(GOGET) -u github.com/raviqqe/liche