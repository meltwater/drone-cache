
# drone-cache 
[![semver](https://img.shields.io/badge/semver-1.0.0-blue.svg?cacheSeconds=2592000)](https://github.com/meltwater/drone-cache/releases) [![Maintenance](https://img.shields.io/maintenance/yes/2019.svg)](https://github.com/meltwater/drone-cache/commits/master) [![Drone](https://drone.meltwater.io/api/badges/meltwater/drone-cache/status.svg)](https://drone.meltwater.io/meltwater/drone-cache) [![Go Doc](https://godoc.org/github.com/meltwater/drone-cache?status.svg)](http://godoc.org/github.com/meltwater/drone-cache) [![Go Report Card](https://goreportcard.com/badge/github.com/meltwater/drone-cache)](https://goreportcard.com/report/github.com/meltwater/drone-cache) [![](https://images.microbadger.com/badges/image/meltwater/drone-cache.svg)](https://microbadger.com/images/meltwater/drone-cache) [![](https://images.microbadger.com/badges/version/meltwater/drone-cache.svg)](https://microbadger.com/images/meltwater/drone-cache)

<p align="center"><img src="images/drone_gopher.png" width="400"></p>

Drone plugin for caching artifacts to a S3 bucket or to a mounted volume.
Use this plugin for caching build artifacts to speed up your build times.
This plugin can create and restore caches of any folders.

For the usage information and a list of the available options please take a look at
[usage](#usage) and checkout [examples](#examples). If you want to learn more about custom cache keys, see [cache key templates](docs/cache_key_templates.md).

## Examples

### Drone Configuration examples

> `!!!` The example Yaml configurations in this file are using the legacy 0.8 syntax. If you are using Drone 1.0 or Drone Cloud please ensure you use the appropriate 1.0 syntax. [Learn more here](https://docs.drone.io/config/pipeline/migrating/#plugins).

The following is a sample configuration in your .drone.yml file:

#### Simple

```yaml
pipeline:
  restore-cache:
    image: meltwater/drone-cache
    pull: true
    # backend: "s3" (default)
    restore: true
    bucket: drone-cache-bucket
    region: eu-west-1
    secrets: [aws_access_key_id, aws_secret_access_key]
    mount:
      - 'deps'
      - '_dialyzer'

  deps:
    image: elixir:1.6.5
    pull: true
    commands:
      - mix local.hex --force
      - mix local.rebar --force
      - mix deps.get
      - mix dialyzer --halt-exit-status

rebuild-deps-cache:
    image: meltwater/drone-cache
    pull: true
    # backend: "s3" (default)
    rebuild: true
    bucket: drone-cache-bucket
    region: eu-west-1
    secrets: [aws_access_key_id, aws_secret_access_key]
    mount:
      - 'deps'
```

#### Simple (Filesystem/Volume)

```yaml
pipeline:
  restore-cache:
    image: meltwater/drone-cache
    pull: true
    backend: "filesystem" # (default: s3)
    restore: true
    bucket: drone-cache-bucket
    region: eu-west-1
    secrets: [aws_access_key_id, aws_secret_access_key]
    mount:
      - 'deps'
      - '_dialyzer'
    volumes:
        - '/drone/tmp/cache:/tmp/cache'

  deps:
    image: elixir:1.6.5
    pull: true
    commands:
      - mix local.hex --force
      - mix local.rebar --force
      - mix deps.get
      - mix dialyzer --halt-exit-status

rebuild-deps-cache:
    image: meltwater/drone-cache
    pull: true
    backend: "filesystem" # (default: s3)
    rebuild: true
    bucket: drone-cache-bucket
    region: eu-west-1
    secrets: [aws_access_key_id, aws_secret_access_key]
    mount:
      - 'deps'
    volumes:
        - '/drone/tmp/cache:/tmp/cache'
```

## For more examples see [docs/examples](docs/examples.md)

## Usage

### Using executable (with CLI args)

```console
NAME:
   Drone cache plugin - Drone cache plugin

USAGE:
   drone-cache [global options] command [command options] [arguments...]

VERSION:
   0.9.0

COMMANDS:
     help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --repo.fullname value, --rf value           repository full name [$DRONE_REPO]
   --repo.owner value, --ro value              repository owner [$DRONE_REPO_OWNER]
   --repo.name value, --rn value               repository name [$DRONE_REPO_NAME]
   --repo.link value, --rl value               repository link [$DRONE_REPO_LINK]
   --repo.avatar value, --ra value             repository avatar [$DRONE_REPO_AVATAR]
   --repo.branch value, --rb value             repository default branch [$DRONE_REPO_BRANCH]
   --repo.private, --rp                        repository is private [$DRONE_REPO_PRIVATE]
   --repo.trusted, --rt                        repository is trusted [$DRONE_REPO_TRUSTED]
   --remote.url value, --remu value            git remote url [$DRONE_REMOTE_URL]
   --commit.sha value, --cs value              git commit sha [$DRONE_COMMIT_SHA]
   --commit.ref value, --cr value              git commit ref (default: "refs/heads/master") [$DRONE_COMMIT_REF]
   --commit.branch value, --cb value           git commit branch (default: "master") [$DRONE_COMMIT_BRANCH]
   --commit.message value, --cm value          git commit message [$DRONE_COMMIT_MESSAGE]
   --commit.link value, --cl value             git commit link [$DRONE_COMMIT_LINK]
   --commit.author.name value, --an value      git author name [$DRONE_COMMIT_AUTHOR]
   --commit.author.email value, --ae value     git author email [$DRONE_COMMIT_AUTHOR_EMAIL]
   --commit.author.avatar value, --aa value    git author avatar [$DRONE_COMMIT_AUTHOR_AVATAR]
   --build.event value, --be value             build event (default: "push") [$DRONE_BUILD_EVENT]
   --build.number value, --bn value            build number (default: 0) [$DRONE_BUILD_NUMBER]
   --build.created value, --bc value           build created (default: 0) [$DRONE_BUILD_CREATED]
   --build.started value, --bs value           build started (default: 0) [$DRONE_BUILD_STARTED]
   --build.finished value, --bf value          build finished (default: 0) [$DRONE_BUILD_FINISHED]
   --build.status value, --bstat value         build status (default: "success") [$DRONE_BUILD_STATUS]
   --build.link value, --bl value              build link [$DRONE_BUILD_LINK]
   --build.deploy value, --db value            build deployment target [$DRONE_DEPLOY_TO]
   --yaml.verified, --yv                       build yaml is verified [$DRONE_YAML_VERIFIED]
   --yaml.signed, --ys                         build yaml is signed [$DRONE_YAML_SIGNED]
   --prev.build.number value, --pbn value      previous build number (default: 0) [$DRONE_PREV_BUILD_NUMBER]
   --prev.build.status value, --pbst value     previous build status [$DRONE_PREV_BUILD_STATUS]
   --prev.commit.sha value, --pcs value        previous build sha [$DRONE_PREV_COMMIT_SHA]
   --backend value, -b value                   cache backend to use in plugin (s3, filesystem) (default: "s3") [$PLUGIN_BACKEND]
   --mount value, -m value                     cache directories, an array of folders to cache [$PLUGIN_MOUNT]
   --rebuild, --reb                            rebuild the cache directories [$PLUGIN_REBUILD]
   --restore, --res                            restore the cache directories [$PLUGIN_RESTORE]
   --cache-key value, --chk value              cache key to use for the cache directories [$PLUGIN_CACHE_KEY]
   --archive-format value, --arcfmt value      archive format to use to store the cache directories (tar, gzip) (default: "tar") [$PLUGIN_ARCHIVE_FORMAT]
   --debug value, -d value                     debug [$PLUGIN_DEBUG, $ DEBUG]
   --filesystem-cache-root value, --fcr value  local filesystem root directory for the filesystem cache (default: "/tmp/cache") [$PLUGIN_FILESYSTEM_CACHE_ROOT, $ FILESYSTEM_CACHE_ROOT]
   --endpoint value, -e value                  endpoint for the s3 connection [$PLUGIN_ENDPOINT, $S3_ENDPOINT]
   --access-key value, --akey value            AWS access key [$PLUGIN_ACCESS_KEY, $AWS_ACCESS_KEY_ID, $CACHE_AWS_ACCESS_KEY_ID]
   --secret-key value, --skey value            AWS secret key [$PLUGIN_SECRET_KEY, $AWS_SECRET_ACCESS_KEY, $CACHE_AWS_SECRET_ACCESS_KEY]
   --bucket value, --bckt value                AWS bucket name [$PLUGIN_BUCKET, $S3_BUCKET]
   --region value, --reg value                 AWS bucket region. (us-east-1, eu-west-1, ...) [$PLUGIN_REGION, $S3_REGION]
   --path-style, --ps                          use path style for bucket paths. (true for minio, false for aws) [$PLUGIN_PATH_STYLE]
   --acl value                                 upload files with acl (private, public-read, ...) (default: "private") [$PLUGIN_ACL]
   --encryption value, --enc value             server-side encryption algorithm, defaults to none. (AES256, aws:kms) [$PLUGIN_ENCRYPTION]
   --help, -h                                  show help
   --version, -v                               print the version
```

### Using Docker (with Environment variables)

```bash
$ docker run --rm \
      -e DRONE_REPO=octocat/hello-world \
      -e DRONE_REPO_BRANCH=master \
      -e DRONE_COMMIT_BRANCH=master \
      -e PLUGIN_MOUNT=node_modules \
      -e PLUGIN_RESTORE=false \
      -e PLUGIN_REBUILD=true \
      -e PLUGIN_BUCKET=<bucket> \
      -e AWS_ACCESS_KEY_ID=<token> \
      -e AWS_SECRET_ACCESS_KEY=<secret> \
      meltwater/drone-cache
```

## Development

### Local setup

```console
$ ./scripts/setup_dev_environment.sh
> Done.
```

### Tests

```console
$ ./test
> ...
```

OR

```console
$ docker-compose up -d
> ...
$ go test ./..
> ...
```

### Build Binary

Build the binary with the following commands:

```console
$ make build
> ...
$ go build .
> ...
```

### Build Docker

Build the docker image with the following commands:

```console
$ make docker-build
> ...
```

## Contributing

Please read [CONTRIBUTING.md](CONTRIBUTING.md) for details on our code of conduct, and the process for submitting pull requests to us.

## Versioning

We use [SemVer](http://semver.org/) for versioning. For the versions available, see the [tags on this repository](https://github.com/meltwater/drone-cache/tags).

## Authors and Acknowledgement

* [@dim](https://github.com/dim) Thanks for original work!
* [@kakkoyun](https://github.com/kakkoyun)
* [@salimane](https://github.com/salimane)

> **Special thanks to [@AdamGlazerMW](https://github.com/AdamGlazerMW) for amazing artwork!**

Check out for [all contributors](https://github.com/meltwater/drone-cache/graphs/contributors).

## License and Copyright

This project is licensed under the Apache License 2.0 - see the [LICENSE](LICENSE) file for details
