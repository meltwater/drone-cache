# drone-s3-cache

[![Go Doc](https://godoc.org/github.com/meltwater/drone-s3-cache?status.svg)](http://godoc.org/github.com/meltwater/drone-s3-cache)

Drone plugin for caching artifacts to a S3 bucket. For the
usage information and a listing of the available options please take a look at
[the docs](DOCS.md).

## Build

Build the binary with the following commands:

```
go build
go test
```

## Docker

Build the docker image with the following commands:

```
make drone-s3-cache
make docker.build
```

Please note incorrectly building the image for the correct x64 linux and with
GCO disabled will result in an error when running the Docker image:

```
docker: Error response from daemon: Container command
'/bin/drone-s3-cache' not found or does not exist..
```

## Usage

Execute from the working directory:

```
docker run --rm \
  -e DRONE_REPO=octocat/hello-world \
  -e DRONE_REPO_BRANCH=master \
  -e DRONE_COMMIT_BRANCH=master \
  -e PLUGIN_MOUNT=node_modules \
  -e PLUGIN_RESTORE=false \
  -e PLUGIN_REBUILD=true \
  -e PLUGIN_BUCKET=<bucket> \
  -e AWS_ACCESS_KEY_ID=<token> \
  -e AWS_SECRET_ACCESS_KEY=<secret> \
  meltwater/drone-s3-cache
```
