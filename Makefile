VERSION := $(strip $(shell [ -d .git ] && git describe --always --tags --dirty))
BUILD_DATE := $(shell date -u +"%Y-%m-%dT%H:%M:%S%Z")
VCS_REF := $(strip $(shell [ -d .git ] && git rev-parse --short HEAD))

GO_PACKAGES=$(shell go list ./... | grep -v -E '/vendor/|/test')
GO_FILES:=$(shell find . -name \*.go -print)
GOPATH:=$(firstword $(subst :, ,$(shell go env GOPATH)))

GOLANGCI_LINT_VERSION=v1.21.0
GOLANGCI_LINT_BIN=$(GOPATH)/bin/golangci-lint
EMBEDMD_BIN=$(GOPATH)/bin/embedmd
GOTEST_BIN=$(GOPATH)/bin/gotest
GORELEASER_VERSION=v0.120
GORELEASER_BIN=$(GOPATH)/bin/goreleaser
LICHE_BIN=$(GOPATH)/bin/liche

.PHONY: default all
default: drone-cache
all: drone-cache

.PHONY: setup
setup:
	./scripts/setup_dev_environment.sh

drone-cache: vendor main.go $(wildcard *.go) $(wildcard */*.go)
	CGO_ENABLED=0 go build -mod=vendor -a -tags netgo -ldflags '-s -w -X main.version=$(VERSION)' -o $@ .

.PHONY: build
build: main.go $(wildcard *.go) $(wildcard */*.go)
	go build -mod=vendor -tags netgo -ldflags '-X main.version=$(VERSION)' -o drone-cache .

.PHONY: release
release: drone-cache $(GORELEASER_BIN)
	${GORELEASER_BIN} release --rm-dist

.PHONY: snapshot
snapshot: drone-cache $(GORELEASER_BIN)
	${GORELEASER_BIN} release --skip-publish --rm-dist --snapshot

.PHONY: clean
clean:
	rm -f drone-cache
	rm -rf target

tmp/help.txt: drone-cache
	mkdir -p tmp
	./drone-cache --help &> tmp/help.txt

README.md: tmp/help.txt
	${EMBEDMD_BIN} -w README.md

tmp/docs.txt: drone-cache
	@echo "IMPLEMENT ME"

DOCS.md: tmp/docs.txt
	${EMBEDMD_BIN} -w DOCS.md

docs: clean README.md DOCS.md ${LICHE_BIN}
	@$(LICHE_BIN) --recursive docs --document-root .
	@$(LICHE_BIN) --exclude "(goreportcard.com)" --document-root . *.md

.PHONY: vendor
vendor:
	@go mod tidy
	@go mod vendor -v

.PHONY: compress
compress: drone-cache
	# Add as dependency
	@upx drone-cache

.PHONY: container
container: release Dockerfile
	@docker build --build-arg BUILD_DATE="$(BUILD_DATE)" \
		--build-arg VERSION="$(VERSION)" \
		--build-arg VCS_REF="$(VCS_REF)" \
		--build-arg DOCKERFILE_PATH="/Dockerfile" \
		-t meltwater/drone-cache:latest .

.PHONY: container-dev
container-dev: snapshot Dockerfile
	@docker build --build-arg BUILD_DATE="$(BUILD_DATE)" \
		--build-arg VERSION="$(VERSION)" \
		--build-arg VCS_REF="$(VCS_REF)" \
		--build-arg DOCKERFILE_PATH="/Dockerfile" \
		--no-cache \
		-t meltwater/drone-cache:dev .

.PHONY: container-push
container-push: container
	docker push meltwater/drone-cache:latest

.PHONY: container-push-dev
container-push-dev: container-dev
	docker push meltwater/drone-cache:dev

.PHONY: test
test: $(GOTEST_BIN)
	docker-compose up -d && sleep 1
	-$(GOTEST_BIN) -race -short -cover -failfast -tags=integration ./...
	docker-compose down -v

.PHONY: test-integration
test-integration: $(GOTEST_BIN)
	docker-compose up -d && sleep 1
	-$(GOTEST_BIN) -race -cover -tags=integration -v ./...
	docker-compose down -v

.PHONY: test-unit
test-unit: $(GOTEST_BIN)
	$(GOTEST_BIN) -race -cover -benchmem -v ./...

.PHONY: test-e2e
test-e2e: $(GOTEST_BIN)
	docker-compose up -d && sleep 1
	-$(GOTEST_BIN) -race -cover -tags=integration -v ./internal/plugin
	docker-compose down -v

.PHONY: lint
lint: $(GOLANGCI_LINT_BIN)
	# Check .golangci.yml for configuration
	$(GOLANGCI_LINT_BIN) run -v --enable-all --skip-dirs tmp -c .golangci.yml

.PHONY: fix
fix: $(GOLANGCI_LINT_BIN) format
	$(GOLANGCI_LINT_BIN) run --fix --enable-all --skip-dirs tmp -c .golangci.yml

.PHONY: format
format:
	@gofmt -w -s $(GO_FILES)

$(GOTEST_BIN):
	GO111MODULE=off go get -u github.com/rakyll/gotest

$(EMBEDMD_BIN):
	GO111MODULE=off go get -u github.com/campoy/embedmd

$(GOLANGCI_LINT_BIN):
	curl -sfL https://raw.githubusercontent.com/golangci/golangci-lint/$(GOLANGCI_LINT_VERSION)/install.sh \
		| sed -e '/install -d/d' \
		| sh -s -- -b $(GOPATH)/bin $(GOLANGCI_LINT_VERSION)

$(GORELEASER_BIN):
	curl -sfL https://install.goreleaser.com/github.com/goreleaser/goreleaser.sh \
		| VERSION=${GORELEASER_VERSION} sh -s -- -b $(GOPATH)/bin $(GORELEASER_BIN)

${LICHE_BIN}:
	GO111MODULE=on go get -u github.com/raviqqe/liche
