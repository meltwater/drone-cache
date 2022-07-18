# drone-cache

[![Latest Release](https://img.shields.io/github/release/meltwater/drone-cache.svg?)](https://github.com/meltwater/drone-cache/releases/latest) [![Maintenance](https://img.shields.io/maintenance/yes/2022.svg)](https://github.com/meltwater/drone-cache/commits/master) ![GitHub](https://img.shields.io/github/license/meltwater/drone-cache) [![drone](https://cloud.drone.io/api/badges/meltwater/drone-cache/status.svg)](https://cloud.drone.io/meltwater/drone-cache) ![release](https://github.com/meltwater/drone-cache/workflows/release/badge.svg) ![snapshot](https://github.com/meltwater/drone-cache/workflows/snapshot/badge.svg)

[![Go Doc](https://godoc.org/github.com/meltwater/drone-cache?status.svg)](http://godoc.org/github.com/meltwater/drone-cache) [![Go Code reference](https://img.shields.io/badge/code%20reference-go.dev-darkblue.svg)](https://pkg.go.dev/github.com/meltwater/drone-cache?tab=subdirectories) [![CII Best Practices](https://bestpractices.coreinfrastructure.org/projects/2713/badge)](https://bestpractices.coreinfrastructure.org/projects/2713) [![Go Report Card](https://goreportcard.com/badge/github.com/meltwater/drone-cache)](https://goreportcard.com/report/github.com/meltwater/drone-cache) [![codebeat badge](https://codebeat.co/badges/802c6149-ac2d-4514-8648-f618c63a8d9e)](https://codebeat.co/projects/github-com-meltwater-drone-cache-master)

[![meltwater/drone-cache on DockerHub](https://img.shields.io/badge/docker-ready-blue.svg)](https://hub.docker.com/r/meltwater/drone-cache) [![DockerHub Pulls](https://img.shields.io/docker/pulls/meltwater/drone-cache.svg)](https://hub.docker.com/r/meltwater/drone-cache)

<p align="center"><img src="images/drone_gopher.png" width="400"></p>

A Drone plugin for caching current workspace files between builds to reduce your build times. `drone-cache` is a small CLI program, written in Go without any external OS dependencies (such as tar, etc).

With `drone-cache`, you can provide your **own cache key templates**, specify **archive format** (tar, tar.gz, etc) and you can use [**popular object storage**](#supported-storage-backends) as storage for your cached files, even better you can implement **your custom storage backend** to cover your use case.

For detailed usage information and a list of available options please take a look at [usage](#usage) and [examples](#example-usage-of-drone-cache). If you want to learn more about custom cache keys, see [cache key templates](docs/cache_key_templates.md).

If you want to learn more about the story behind `drone-cache`, you can read our blogpost [Making Drone Builds 10 Times Faster!](https://underthehood.meltwater.com/blog/2019/04/10/making-drone-builds-10-times-faster/)!

## Supported Storage Backends

* [AWS S3](https://aws.amazon.com/s3/)
  * [Configuration](#)
  * [Example](#)
  * Other AWS API compatible stores:
    * [Minio](https://min.io/)
    * [Red Hat Ceph](https://www.redhat.com/en/technologies/storage/ceph)
    * [IBM Object Store](https://www.ibm.com/cloud/object-storage)
    * and many many others
* [Azure Storage](https://azure.microsoft.com/en-us/services/storage/blobs/)
  * [Configuration](#)
  * [Example](#)
* [Google Cloud Storage](https://cloud.google.com/storage/)
  * [Configuration](#)
  * [Example](#)
* or any mounted local volume
  * [Configuration](#)
  * [Example](#)

## How does it work

`drone-cache` stores mounted directories and files under a key at the specified backend (by default S3).

Use this plugin to cache data that makes your builds faster. In the case of a _cache miss_ or an _empty cache_ restore it will fail silently in won't break your running pipeline.

The best example would be to use this with your package managers such as Mix, Bundler or Maven. After your initial download, you can build a cache and then you can restore that cache in your next build.

<p align="center"><img src="images/diagram.png" width="400"></p>

With restored dependencies from a cache, commands like `mix deps.get` will only need to download new dependencies, rather than re-download every package on each build.

## Example Usage of drone-cache

The following example configuration file (`.drone.yml`) shows the most common use of drone-cache.

[//]: # (TODO: Move to a dedicated directory in docs, per backend!)
### Simple (with AWS S3 backend)

```yaml
kind: pipeline
name: default

steps:
  - name: restore-cache
    image: meltwater/drone-cache
    environment:
      AWS_ACCESS_KEY_ID:
        from_secret: aws_access_key_id
      AWS_SECRET_ACCESS_KEY:
        from_secret: aws_secret_access_key
    pull: true
    settings:
      restore: true
      cache_key: {{ .Commit.Branch }}-{{ checksum "go.mod" }} # default if ommitted is {{ .Commit.Branch }}
      bucket: drone-cache-bucket
      region: eu-west-1
      mount:
        - 'vendor'

  - name: build
    image: golang:1.18.4
    pull: true
    commands:
      - make drone-cache

  - name: rebuild-cache
    image: meltwater/drone-cache
    pull: true
    environment:
      AWS_ACCESS_KEY_ID:
        from_secret: aws_access_key_id
      AWS_SECRET_ACCESS_KEY:
        from_secret: aws_secret_access_key
    settings:
      rebuild: true
      cache_key: {{ .Commit.Branch }}-{{ checksum "go.mod" }} # default if ommitted is {{ .Commit.Branch }}
      bucket: drone-cache-bucket
      region: eu-west-1
      mount:
        - 'vendor'

```

### More Examples

- examples for Drone, see [docs/examples/drone-1.0.md](docs/examples/drone.md)

## Usage

### Using executable (with CLI args)

[embedmd]:# (tmp/help.txt)
```txt
```

### Using Docker (with Environment variables)

```bash
$ docker run --rm \
      -v "$(pwd)":/app \
      -e DRONE_REPO=octocat/hello-world \
      -e DRONE_REPO_BRANCH=master \
      -e DRONE_COMMIT_BRANCH=master \
      -e PLUGIN_MOUNT=/app/node_modules \
      -e PLUGIN_RESTORE=false \
      -e PLUGIN_REBUILD=true \
      -e PLUGIN_BUCKET=<bucket> \
      -e AWS_ACCESS_KEY_ID=<token> \
      -e AWS_SECRET_ACCESS_KEY=<secret> \
      meltwater/drone-cache
```

## Development

[embedmd]:# (tmp/make_help.txt)
```txt
```

## Releases

Release management handled by the CI pipeline. When you create a tag on `master` branch, CI handles the rest.

You can find released artifacts (binaries, code, archives) under [releases](https://github.com/meltwater/drone-cache/releases).

You can find released images at [DockerHub](https://hub.docker.com/r/meltwater/drone-cache/tags).

**PLEASE DO NOT INTRODUCE BREAKING CHANGES**

> Keep in mind that users usually use the image tagged with `latest` in their pipeline, please make sure you do not interfere with their working workflow. Latest stable releases will be tagged with the `latest`.

## Versioning

`drone-cache` uses [SemVer](http://semver.org/) for versioning. For the versions available, see the [tags on this repository](https://github.com/meltwater/drone-cache/tags).

As the versioning scheme dictates, `drone-cache` respects _backward compatibility_ within the major versions. However, the project only offers guarantees regarding the command-line interface (flags and environment variables). **Any exported public package can change its API.**

## Authors and Acknowledgement

See the list of [all contributors](https://github.com/meltwater/drone-cache/graphs/contributors).

- [@kakkoyun](https://github.com/kakkoyun) - Thank you Kemal for bringing drone-cache to life, and building most of the initial version.
- [@AdamGlazerMW](https://github.com/AdamGlazerMW) - Special thanks to Adam for the amazing artwork!
- [@dim](https://github.com/dim) - Thanks for the [original work](https://github.com/bsm/drone-s3-cache) that inspired drone-cache!

### Inspiration

- [github.com/bsm/drone-s3-cache](https://github.com/bsm/drone-s3-cache) (original work)
- [github.com/Drillster/drone-volume-cache](https://github.com/Drillster/drone-volume-cache)
- [github.com/drone/drone-cache-lib](https://github.com/drone/drone-cache-lib)

## Contributing

Please read [CONTRIBUTING.md](CONTRIBUTING.md) to understand how to submit pull requests to us, and also see our [code of conduct](CODE_OF_CONDUCT.md).

## Future work

All ideas for new features and bug reports will be kept in [github.com/meltwater/drone-cache/issues](https://github.com/meltwater/drone-cache/issues).

One bigger area of future investment is to add a couple of [new storage backends](https://github.com/meltwater/drone-cache/labels/storage-backend) for caching the workspace files.

## License and Copyright

This project is licensed under the [Apache License 2.0](LICENSE).
