default: drone-s3-cache
all: drone-s3-cache

drone-s3-cache: fetch-dependecies main.go $(wildcard *.go) $(wildcard */*.go)
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -mod=vendor -a -ldflags '-w' -o $@ .

clean:
	rm -f drone-s3-cache

.PHONY: default all clean

fetch-dependecies:
	go mod vendor

.PHONY: fetch-dependecies

docker-build: Dockerfile
	docker build -t meltwater/drone-s3-cache:latest .

docker-push: docker-build
	docker push meltwater/drone-s3-cache:latest

.PHONY: docker-build docker-push

