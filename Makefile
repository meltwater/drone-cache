PKG=$(shell glide nv)

default: vet test

vet:
	go vet $(PKG)

test:
	go test $(PKG)

all: drone-s3-cache

.PHONY: default vet test all

docker.build: drone-s3-cache Dockerfile
	docker build -t blacksquaremedia/drone-s3-cache:latest .

docker.push: docker.build
	docker push blacksquaremedia/drone:latest

.PHONY: docker.build docker.push

drone-s3-cache: main.go $(wildcard *.go) $(wildcard */*.go)
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -ldflags '-w' -o $@ .

