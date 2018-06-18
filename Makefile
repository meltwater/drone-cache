PKG=$(shell glide nv)

default: vet test

vet:
	go vet $(PKG)

test:
	go test $(PKG)

all: drone-s3-cache

clean:
	rm -f drone-s3-cache

.PHONY: default vet test clean all

install-glide:
	curl https://glide.sh/get | sh

fetch-dependecies: install-glide
	glide install

.PHONY: install-glide fetch-dependecies

docker.build: drone-s3-cache Dockerfile
	docker build -t meltwater/drone-s3-cache:latest .

docker.push: docker.build
	docker push meltwater/drone-s3-cache:latest

.PHONY: docker.build docker.push

drone-s3-cache: main.go $(wildcard *.go) $(wildcard */*.go)
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -ldflags '-w' -o $@ .
