default: drone-cache
all: drone-cache

drone-cache: fetch-dependencies main.go $(wildcard *.go) $(wildcard */*.go)
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -mod=vendor -a -ldflags '-s -w' -o $@ .

clean:
	rm -f drone-cache

.PHONY: default all clean

fetch-dependencies:
	go mod vendor

.PHONY: fetch-dependencies

compress: drone-cache
	upx --brute drone-cache

.PHONY: compress

docker-build: Dockerfile
	docker build -t meltwater/drone-cache:latest .

docker-push: docker-build
	docker push meltwater/drone-cache:latest

.PHONY: docker-build docker-push
