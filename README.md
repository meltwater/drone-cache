# drone-s3-cache

[![Go Doc](https://godoc.org/github.com/meltwater/drone-s3-cache?status.svg)](http://godoc.org/github.com/meltwater/drone-s3-cache)
[![Drone](https://drone.meltwater.io/api/badges/meltwater/drone-s3-cache/status.svg)](https://drone.meltwater.io/meltwater/drone-s3-cache)
[![Maintenance](https://img.shields.io/maintenance/yes/2018.svg)](https://github.com/meltwater/drone-s3-cache/commits/master)

Drone plugin for caching artifacts to a S3 bucket.
Use this plugin for caching build artifacts to speed up your build times.
This plugin can create and restore caches of any folders.

For the usage information and a listing of the available options please take a look at
[usage](#usage).

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
make docker-build
```

Please note incorrectly building the image for the correct x64 linux and with
GCO disabled will result in an error when running the Docker image:

```
docker: Error response from daemon: Container command
'/bin/drone-s3-cache' not found or does not exist..
```

# Usage

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

## Config

The following parameters are used to configure the plugin:

* **endpoint** - custom endpoint URL (optional, to use a S3 compatible non-Amazon service)
* **access_key** - amazon key (optional)
* **secret_key** - amazon secret key (optional)
* **bucket** - bucket name
* **region** - bucket region (`us-east-1`, `eu-west-1`, etc)
* **encryption** - if provided, use server-side encryption (`AES256`, `aws:kms`, etc)
* **acl** - access to files that are uploaded (`private`, `public-read`, etc)
* **path_style** - whether path style URLs should be used (true for minio, false for aws)
* **mount**   - one or an array of folders to cache
* **rebuild** - boolean flag to trigger a rebuild
* **restore** - boolean flag to trigger a restore

The following secret values can be set to configure the plugin.

* **AWS_ACCESS_KEY_ID** or **CACHE_AWS_ACCESS_KEY_ID** - corresponds to **access_key**
* **AWS_SECRET_ACCESS_KEY** or **CACHE_AWS_SECRET_ACCESS_KEY** - corresponds to **secret_key**
* **S3_BUCKET** - corresponds to **bucket**
* **S3_REGION** - corresponds to **region**
* **PLUGIN_ENDPOINT** - corresponds to **endpoint**

## Example

The following is a sample configuration in your .drone.yml file:

```yaml

pipeline:
  s3_cache_restore:
    bucket: my-drone-bucket
    image: meltwater/drone-s3-cache
    restore: true
    mount:
    - node_modules

  build:
    image: node:latest
    commands:
    - npm install

  s3_cache_rebuild:
    bucket: my-drone-bucket
    image: meltwater/drone-s3-cache
    rebuild: true
    mount:
    - node_modules

```
