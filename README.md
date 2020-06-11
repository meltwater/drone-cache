# drone-cache

[![Latest Release](https://img.shields.io/github/release/meltwater/drone-cache.svg?)](https://github.com/meltwater/drone-cache/releases/latest) [![Maintenance](https://img.shields.io/maintenance/yes/2020.svg)](https://github.com/meltwater/drone-cache/commits/master) ![GitHub](https://img.shields.io/github/license/meltwater/drone-cache) [![Gitter](https://badges.gitter.im/drone-cache/community.svg)](https://gitter.im/drone-cache/community?utm_source=badge&utm_medium=badge&utm_campaign=pr-badge) [![drone](https://cloud.drone.io/api/badges/meltwater/drone-cache/status.svg)](https://cloud.drone.io/meltwater/drone-cache) ![release](https://github.com/meltwater/drone-cache/workflows/release/badge.svg) ![snapshot](https://github.com/meltwater/drone-cache/workflows/snapshot/badge.svg)

[![Go Doc](https://godoc.org/github.com/meltwater/drone-cache?status.svg)](http://godoc.org/github.com/meltwater/drone-cache) [![Go Code reference](https://img.shields.io/badge/code%20reference-go.dev-darkblue.svg)](https://pkg.go.dev/github.com/meltwater/drone-cache?tab=subdirectories) [![CII Best Practices](https://bestpractices.coreinfrastructure.org/projects/2713/badge)](https://bestpractices.coreinfrastructure.org/projects/2713) [![Go Report Card](https://goreportcard.com/badge/github.com/meltwater/drone-cache)](https://goreportcard.com/report/github.com/meltwater/drone-cache) [![codebeat badge](https://codebeat.co/badges/802c6149-ac2d-4514-8648-f618c63a8d9e)](https://codebeat.co/projects/github-com-meltwater-drone-cache-master)

[![meltwater/drone-cache on DockerHub](https://img.shields.io/badge/docker-ready-blue.svg)](https://hub.docker.com/r/meltwater/drone-cache) [![DockerHub Pulls](https://img.shields.io/docker/pulls/meltwater/drone-cache.svg)](https://hub.docker.com/r/meltwater/drone-cache) [![](https://images.microbadger.com/badges/image/meltwater/drone-cache.svg)](https://microbadger.com/images/meltwater/drone-cache) [![](https://images.microbadger.com/badges/version/meltwater/drone-cache.svg)](https://microbadger.com/images/meltwater/drone-cache)

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
    image: golang:1.14.4-alpine3.12
    pull: true
    commands:
      - apk add --update make git
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
NAME:
   Drone cache plugin - Drone cache plugin

USAGE:
   drone-cache [global options] command [command options] [arguments...]

VERSION:
   v1.1.0-rc0

COMMANDS:
   help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --log.level value                     log filtering level. ('error', 'warn', 'info', 'debug') (default: "info") [$PLUGIN_LOG_LEVEL, $LOG_LEVEL]
   --log.format value                    log format to use. ('logfmt', 'json') (default: "logfmt") [$PLUGIN_LOG_FORMAT, $LOG_FORMAT]
   --repo.fullname value                 repository full name [$DRONE_REPO]
   --repo.namespace value                repository namespace [$DRONE_REPO_NAMESPACE]
   --repo.owner value                    repository owner (for Drone version < 1.0) [$DRONE_REPO_OWNER]
   --repo.name value                     repository name [$DRONE_REPO_NAME]
   --repo.link value                     repository link [$DRONE_REPO_LINK]
   --repo.avatar value                   repository avatar [$DRONE_REPO_AVATAR]
   --repo.branch value                   repository default branch [$DRONE_REPO_BRANCH]
   --repo.private                        repository is private (default: false) [$DRONE_REPO_PRIVATE]
   --repo.trusted                        repository is trusted (default: false) [$DRONE_REPO_TRUSTED]
   --remote.url value                    git remote url [$DRONE_REMOTE_URL]
   --commit.sha value                    git commit sha [$DRONE_COMMIT_SHA]
   --commit.ref value                    git commit ref (default: "refs/heads/master") [$DRONE_COMMIT_REF]
   --commit.branch value                 git commit branch (default: "master") [$DRONE_COMMIT_BRANCH]
   --commit.message value                git commit message [$DRONE_COMMIT_MESSAGE]
   --commit.link value                   git commit link [$DRONE_COMMIT_LINK]
   --commit.author.name value            git author name [$DRONE_COMMIT_AUTHOR]
   --commit.author.email value           git author email [$DRONE_COMMIT_AUTHOR_EMAIL]
   --commit.author.avatar value          git author avatar [$DRONE_COMMIT_AUTHOR_AVATAR]
   --build.event value                   build event (default: "push") [$DRONE_BUILD_EVENT]
   --build.number value                  build number (default: 0) [$DRONE_BUILD_NUMBER]
   --build.created value                 build created (default: 0) [$DRONE_BUILD_CREATED]
   --build.started value                 build started (default: 0) [$DRONE_BUILD_STARTED]
   --build.finished value                build finished (default: 0) [$DRONE_BUILD_FINISHED]
   --build.status value                  build status (default: "success") [$DRONE_BUILD_STATUS]
   --build.link value                    build link [$DRONE_BUILD_LINK]
   --build.deploy value                  build deployment target [$DRONE_DEPLOY_TO]
   --yaml.verified                       build yaml is verified (default: false) [$DRONE_YAML_VERIFIED]
   --yaml.signed                         build yaml is signed (default: false) [$DRONE_YAML_SIGNED]
   --prev.build.number value             previous build number (default: 0) [$DRONE_PREV_BUILD_NUMBER]
   --prev.build.status value             previous build status [$DRONE_PREV_BUILD_STATUS]
   --prev.commit.sha value               previous build sha [$DRONE_PREV_COMMIT_SHA]
   --backend value                       cache backend to use in plugin (s3, filesystem, sftp, azure, gcs) (default: "s3") [$PLUGIN_BACKEND]
   --mount value                         cache directories, an array of folders to cache [$PLUGIN_MOUNT]
   --rebuild                             rebuild the cache directories (default: false) [$PLUGIN_REBUILD]
   --restore                             restore the cache directories (default: false) [$PLUGIN_RESTORE]
   --cache-key value                     cache key to use for the cache directories [$PLUGIN_CACHE_KEY]
   --remote-root value                   remote root directory to contain all the cache files created (default repo.name) [$PLUGIN_REMOTE_ROOT]
   --local-root value                    local root directory to base given mount paths (default pwd [present working directory]) [$PLUGIN_LOCAL_ROOT]
   --override                            override even if cache key already exists in backend (default: true) [$PLUGIN_OVERRIDE]
   --archive-format value                archive format to use to store the cache directories (tar, gzip) (default: "tar") [$PLUGIN_ARCHIVE_FORMAT]
   --compression-level value             compression level to use for gzip compression when archive-format specified as gzip
                                             (check https://godoc.org/compress/flate#pkg-constants for available options) (default: -1) [$PLUGIN_COMPRESSION_LEVEL]
   --skip-symlinks                       skip symbolic links in archive (default: false) [$PLUGIN_SKIP_SYMLINKS, $SKIP_SYMLINKS]
   --debug                               debug (default: false) [$PLUGIN_DEBUG, $DEBUG]
   --backend.operation-timeout value     timeout value to use for each storage operations (default: 3m0s) [$PLUGIN_BACKEND_OPERATION_TIMEOUT, $BACKEND_OPERATION_TIMEOUT]
   --endpoint value                      endpoint for the s3/cloud storage connection [$PLUGIN_ENDPOINT, $S3_ENDPOINT, $GCS_ENDPOINT]
   --bucket value                        AWS bucket name [$PLUGIN_BUCKET, $S3_BUCKET, $GCS_BUCKET]
   --filesystem.cache-root value         local filesystem root directory for the filesystem cache (default: "/tmp/cache") [$PLUGIN_FILESYSTEM_CACHE_ROOT, $FILESYSTEM_CACHE_ROOT]
   --access-key value                    AWS access key [$PLUGIN_ACCESS_KEY, $AWS_ACCESS_KEY_ID, $CACHE_AWS_ACCESS_KEY_ID]
   --secret-key value                    AWS secret key [$PLUGIN_SECRET_KEY, $AWS_SECRET_ACCESS_KEY, $CACHE_AWS_SECRET_ACCESS_KEY]
   --region value                        AWS bucket region. (us-east-1, eu-west-1, ...) [$PLUGIN_REGION, $S3_REGION]
   --path-style                          AWS path style to use for bucket paths. (true for minio, false for aws) (default: false) [$PLUGIN_PATH_STYLE, $AWS_PLUGIN_PATH_STYLE]
   --acl value                           upload files with acl (private, public-read, ...) (default: "private") [$PLUGIN_ACL, $AWS_ACL]
   --encryption value                    server-side encryption algorithm, defaults to none. (AES256, aws:kms) [$PLUGIN_ENCRYPTION, $AWS_ENCRYPTION]
   --gcs.api-key value                   Google service account API key [$PLUGIN_API_KEY, $GCP_API_KEY]
   --gcs.json-key value                  Google service account JSON key [$PLUGIN_JSON_KEY, $GCS_CACHE_JSON_KEY]
   --gcs.acl value                       upload files with acl (private, public-read, ...) (default: "private") [$PLUGIN_GCS_ACL, $GCS_ACL]
   --gcs.encryption-key value            server-side encryption key, must be a 32-byte AES-256 key, defaults to none
                                             (See https://cloud.google.com/storage/docs/encryption for details.) [$PLUGIN_GCS_ENCRYPTION_KEY, $GCS_ENCRYPTION_KEY]
   --azure.account-name value            Azure Blob Storage Account Name [$PLUGIN_ACCOUNT_NAME, $AZURE_ACCOUNT_NAME]
   --azure.account-key value             Azure Blob Storage Account Key [$PLUGIN_ACCOUNT_KEY, $AZURE_ACCOUNT_KEY]
   --azure.blob-container-name value     Azure Blob Storage container name [$PLUGIN_CONTAINER, $AZURE_CONTAINER_NAME]
   --azure.blob-storage-url value        Azure Blob Storage URL (default: "blob.core.windows.net") [$AZURE_BLOB_STORAGE_URL]
   --azure.blob-max-retry-requets value  Azure Blob Storage Max Retry Requests (default: 4) [$AZURE_BLOB_MAX_RETRY_REQUESTS]
   --sftp.cache-root value               sftp root directory [$SFTP_CACHE_ROOT]
   --sftp.username value                 sftp username [$PLUGIN_USERNAME, $SFTP_USERNAME]
   --sftp.password value                 sftp password [$PLUGIN_PASSWORD, $SFTP_PASSWORD]
   --sftp.public-key-file value          sftp public key file path [$PLUGIN_PUBLIC_KEY_FILE, $SFTP_PUBLIC_KEY_FILE]
   --sftp.auth-method value              sftp auth method, defaults to none. (PASSWORD, PUBLIC_KEY_FILE) [$SFTP_AUTH_METHOD]
   --sftp.host value                     sftp host [$SFTP_HOST]
   --sftp.port value                     sftp port [$SFTP_PORT]
   --help, -h                            show help (default: false)
   --version, -v                         print the version (default: false)
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
Usage:
  make <target>

Targets:
  setup          	  Setups dev environment
  drone-cache    	  Runs drone-cache target
  build          	  Runs build target, always builds
  clean          	  Cleans build resourcess
  docs           	  Generates docs
  generate       	  Generate documentation, website and yaml files,
  vendor         	  Updates vendored copy of dependencies
  compress       	  Creates compressed binary
  container      	  Builds drone-cache docker image with latest tag
  container-push 	  Pushes latest $(CONTAINER_REPO) image to repository
  test           	  Runs tests
  test-integration	  Runs integration tests
  test-unit      	  Runs unit tests
  lint           	  Runs golangci-lint analysis
  fix            	  Runs golangci-lint fix
  format         	  Runs gofmt
  help           	  Shows this help message
```

```txt
Usage:
  make <target>

Targets:
  setup          	  Setups dev environment
  drone-cache    	  Runs drone-cache target
  build          	  Runs build target, always builds
  clean          	  Cleans build resourcess
  docs           	  Generates docs
  generate       	  Generate documentation, website and yaml files,
  vendor         	  Updates vendored copy of dependencies
  compress       	  Creates compressed binary
  container      	  Builds drone-cache docker image with latest tag
  container-push 	  Pushes latest $(CONTAINER_REPO) image to repository
  test           	  Runs tests
  test-integration	  Runs integration tests
  test-unit      	  Runs unit tests
  lint           	  Runs golangci-lint analysis
  fix            	  Runs golangci-lint fix
  format         	  Runs gofmt
  help           	  Shows this help message
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

- [@dim](https://github.com/dim) - Thanks for [original work](https://github.com/bsm/drone-s3-cache)!
- [@AdamGlazerMW](https://github.com/AdamGlazerMW) - Special thanks to Adam for the amazing artwork!

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
