VERSION := $(strip $(shell [ -d .git ] && git describe --always --tags --dirty))
BUILD_DATE := $(shell date -u +"%Y-%m-%dT%H:%M:%S%Z")
VCS_REF := $(strip $(shell [ -d .git ] && git rev-parse --short HEAD))

default: drone-cache
all: drone-cache

drone-cache: fetch-dependencies main.go $(wildcard *.go) $(wildcard */*.go)
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -mod=vendor -a -ldflags '-s -w' -o $@ .

build: fetch-dependencies main.go $(wildcard *.go) $(wildcard */*.go)
	go build -mod=vendor -a -ldflags '-s -w' -o drone-cache .

release:
	goreleaser release --rm-dist

snapshot:
	goreleaser release --skip-publish --rm-dist --snapshot

clean:
	rm -f drone-cache
	rm -rf target

.PHONY: default all clean release snapshot

fetch-dependencies:
	@go mod vendor -v

.PHONY: fetch-dependencies

build-compressed: drone-cache
	@upx drone-cache

.PHONY: build-compressed

docker-build: Dockerfile
	docker build --build-arg BUILD_DATE="$(BUILD_DATE)" \
		--build-arg VERSION="$(VERSION)" \
		--build-arg VCS_REF="$(VCS_REF)" \
		--build-arg DOCKERFILE_PATH="/Dockerfile" \
		-t meltwater/drone-cache:latest .

docker-build-dev: Dockerfile
	docker build --build-arg BUILD_DATE="$(BUILD_DATE)" \
		--build-arg VERSION="$(VERSION)" \
		--build-arg VCS_REF="$(VCS_REF)" \
		--build-arg DOCKERFILE_PATH="/Dockerfile" \
		-t meltwater/drone-cache:dev .

docker-push: docker-build
	docker push meltwater/drone-cache:latest

docker-push-dev: docker-build-dev
	docker push meltwater/drone-cache:dev

.PHONY: docker-build docker-push
