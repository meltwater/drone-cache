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

.PHONY: default all
default: drone-cache
all: drone-cache

drone-cache: vendor main.go $(wildcard *.go) $(wildcard */*.go)
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -mod=vendor -a -ldflags '-s -w -X main.version=$(VERSION)' -o $@ .

.PHONY: build
build: vendor main.go $(wildcard *.go) $(wildcard */*.go)
	go build -mod=vendor -a -ldflags '-s -w -X main.version=$(VERSION)' -o drone-cache .

.PHONY: release
release: build
	goreleaser release --rm-dist

.PHONY: snapshot
snapshot:
	goreleaser release --skip-publish --rm-dist --snapshot

.PHONY: clean
clean:
	rm -f drone-cache
	rm -rf target

tmp/help.txt: clean build
	mkdir -p tmp
	./drone-cache --help &> tmp/help.txt

README.md: tmp/help.txt
	embedmd -w README.md

.PHONY: vendor
vendor:
	@go mod vendor -v

.PHONY: compress
compress: drone-cache
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
	docker-compose up -d
	mkdir -p ./testcache/cache
	gotest -race -short -cover ./...

.PHONY: lint
lint: $(GOLANGCI_LINT_BIN)
	# Check .golangci.yml for configuration
	$(GOLANGCI_LINT_BIN) run -v --enable-all -c .golangci.yml

.PHONY: fix
fix: $(GOLANGCI_LINT_BIN) format
	$(GOLANGCI_LINT_BIN) run --fix --enable-all -c .golangci.yml

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
