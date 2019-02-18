# drone-cache

[![Maintenance](https://img.shields.io/maintenance/yes/2019.svg)](https://github.com/meltwater/drone-cache/commits/master)
[![Drone](https://drone.meltwater.io/api/badges/meltwater/drone-cache/status.svg)](https://drone.meltwater.io/meltwater/drone-cache)
[![Go Doc](https://godoc.org/github.com/meltwater/drone-cache?status.svg)](http://godoc.org/github.com/meltwater/drone-cache)
[![Go Report Card](https://goreportcard.com/badge/github.com/meltwater/drone-cache)](https://goreportcard.com/report/github.com/meltwater/drone-cache)
[![](https://images.microbadger.com/badges/image/meltwater/drone-cache.svg)](https://microbadger.com/images/meltwater/drone-cache "Get your own image badge on microbadger.com")
[![](https://images.microbadger.com/badges/version/meltwater/drone-cache.svg)](https://microbadger.com/images/meltwater/drone-cache "Get your own version badge on microbadger.com")

Drone plugin for caching artifacts to a S3 bucket (or soon to a mounted volume).
Use this plugin for caching build artifacts to speed up your build times.
This plugin can create and restore caches of any folders.

For the usage information and a listing of the available options please take a look at
[usage](#usage) and [examples](#examples).

## Examples

### Drone Configuration examples

The following is a sample configuration in your .drone.yml file:

#### Simple

```yaml
pipeline:
  restore-cache:
    image: meltwater/drone-cache
    pull: true
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
    rebuild: true
    bucket: drone-cache-bucket
    region: eu-west-1
    secrets: [aws_access_key_id, aws_secret_access_key]
    mount:
      - 'deps'
```

#### With custom cache key prefix template

See [cache key templates](#cache-key-templates) section for further information and to learn about syntax.

```yaml
pipeline:
  restore-cache:
    image: meltwater/drone-cache
    pull: true
    restore: true
    cache_key: "{{ .Repo.Name }}_{{ .Commit.Branch }}_{{ .Build.Number }}"
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
    rebuild: true
    cache_key: "{{ .Repo.Name }}_{{ .Commit.Branch }}_{{ .Build.Number }}"
    bucket: drone-cache-bucket
    region: eu-west-1
    secrets: [aws_access_key_id, aws_secret_access_key]
    mount:
      - 'deps'
```

#### With gzip compression

```yaml
pipeline:
  restore-cache:
    image: meltwater/drone-cache
    pull: true
    restore: true
    cache_key: "{{ .Repo.Name }}_{{ .Commit.Branch }}_{{ .Build.Number }}"
    archive_format: "gzip"
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
    rebuild: true
    cache_key: "{{ .Repo.Name }}_{{ .Commit.Branch }}_{{ .Build.Number }}"
    archive_format: "gzip"
    bucket: drone-cache-bucket
    region: eu-west-1
    secrets: [aws_access_key_id, aws_secret_access_key]
    mount:
      - 'deps'
```

#### Debug

```yaml
pipeline:
  restore-cache:
    image: meltwater/drone-cache
    pull: true
    restore: true
    debug: true
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
    rebuild: true
    debug: true
    bucket: drone-cache-bucket
    region: eu-west-1
    secrets: [aws_access_key_id, aws_secret_access_key]
    mount:
      - 'deps'
```

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
   --repo.fullname value, --rf value         repository full name [$DRONE_REPO]
   --repo.owner value, --ro value            repository owner [$DRONE_REPO_OWNER]
   --repo.name value, --rn value             repository name [$DRONE_REPO_NAME]
   --repo.link value, --rl value             repository link [$DRONE_REPO_LINK]
   --repo.avatar value, --ra value           repository avatar [$DRONE_REPO_AVATAR]
   --repo.branch value, --rb value           repository default branch [$DRONE_REPO_BRANCH]
   --repo.private, --rp                      repository is private [$DRONE_REPO_PRIVATE]
   --repo.trusted, --rt                      repository is trusted [$DRONE_REPO_TRUSTED]
   --remote.url value, --remu value          git remote url [$DRONE_REMOTE_URL]
   --commit.sha value, --cs value            git commit sha [$DRONE_COMMIT_SHA]
   --commit.ref value, --cr value            git commit ref (default: "refs/heads/master") [$DRONE_COMMIT_REF]
   --commit.branch value, --cb value         git commit branch (default: "master") [$DRONE_COMMIT_BRANCH]
   --commit.message value, --cm value        git commit message [$DRONE_COMMIT_MESSAGE]
   --commit.link value, --cl value           git commit link [$DRONE_COMMIT_LINK]
   --commit.author.name value, --an value    git author name [$DRONE_COMMIT_AUTHOR]
   --commit.author.email value, --ae value   git author email [$DRONE_COMMIT_AUTHOR_EMAIL]
   --commit.author.avatar value, --aa value  git author avatar [$DRONE_COMMIT_AUTHOR_AVATAR]
   --build.event value, --be value           build event (default: "push") [$DRONE_BUILD_EVENT]
   --build.number value, --bn value          build number (default: 0) [$DRONE_BUILD_NUMBER]
   --build.created value, --bc value         build created (default: 0) [$DRONE_BUILD_CREATED]
   --build.started value, --bs value         build started (default: 0) [$DRONE_BUILD_STARTED]
   --build.finished value, --bf value        build finished (default: 0) [$DRONE_BUILD_FINISHED]
   --build.status value, --bstat value       build status (default: "success") [$DRONE_BUILD_STATUS]
   --build.link value, --bl value            build link [$DRONE_BUILD_LINK]
   --build.deploy value, --db value          build deployment target [$DRONE_DEPLOY_TO]
   --yaml.verified, --yv                     build yaml is verified [$DRONE_YAML_VERIFIED]
   --yaml.signed, --ys                       build yaml is signed [$DRONE_YAML_SIGNED]
   --prev.build.number value, --pbn value    previous build number (default: 0) [$DRONE_PREV_BUILD_NUMBER]
   --prev.build.status value, --pbst value   previous build status [$DRONE_PREV_BUILD_STATUS]
   --prev.commit.sha value, --pcs value      previous build sha [$DRONE_PREV_COMMIT_SHA]
   --mount value, -m value                   cache directories, an array of folders to cache [$PLUGIN_MOUNT]
   --rebuild, --reb                          rebuild the cache directories [$PLUGIN_REBUILD]
   --restore, --res                          restore the cache directories [$PLUGIN_RESTORE]
   --cache-key value, --chk value            cache key to use for the cache directories [$PLUGIN_CACHE_KEY]
   --archive-format value, --arcfmt value    archive format to use to store the cache directories. (tar, gzip) (default: "tar") [$PLUGIN_ARCHIVE_FORMAT]
   --endpoint value, -e value                endpoint for the s3 connection [$PLUGIN_ENDPOINT, $S3_ENDPOINT]
   --access-key value, --akey value          AWS access key [$PLUGIN_ACCESS_KEY, $AWS_ACCESS_KEY_ID, $CACHE_AWS_ACCESS_KEY_ID]
   --secret-key value, --skey value          AWS secret key [$PLUGIN_SECRET_KEY, $AWS_SECRET_ACCESS_KEY, $CACHE_AWS_SECRET_ACCESS_KEY]
   --bucket value, --bckt value              AWS bucket name [$PLUGIN_BUCKET, $S3_BUCKET]
   --region value, --reg value               AWS bucket region. (us-east-1, eu-west-1, ...) [$PLUGIN_REGION, $S3_REGION]
   --path-style, --ps                        use path style for bucket paths. (true for minio, false for aws) [$PLUGIN_PATH_STYLE]
   --acl value                               upload files with acl (private, public-read, ...) (default: "private") [$PLUGIN_ACL]
   --encryption value, --enc value           server-side encryption algorithm, defaults to none. (AES256, aws:kms) [$PLUGIN_ENCRYPTION]
   --help, -h                                show help
   --version, -v                             print the version

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

## Cache Key Templates

`"{{ .Repo.Name }}-{{ .Commit.Branch }}-yadayadayada"`

Cache key template syntax is very basic. You just need to provide a string. In that string you can use variables by prefixing them with a `.` in `{{ }}` construct, from provided metadata object.

Following metadata object is available and pre-populated with current build information for you to use in cache key templates.

For further information about this syntax please see [official docs](https://golang.org/pkg/text/template/) from Go standard library.

```go
{
  Repo {
    Avatar  string "repository avatar [$DRONE_REPO_AVATAR]"
    Branch  string "repository default branch [$DRONE_REPO_BRANCH]"
    Link    string "repository link [$DRONE_REPO_LINK]"
    Name    string "repository name [$DRONE_REPO_NAME]"
    Owner   string "repository owner [$DRONE_REPO_OWNER]"
    Private bool   "repository is private [$DRONE_REPO_PRIVATE]"
    Trusted bool   "repository is trusted [$DRONE_REPO_TRUSTED]"
  }

  Build {
    Created  int    "build created (default: 0) [$DRONE_BUILD_CREATED]"
    Deploy   string "build deployment target [$DRONE_DEPLOY_TO]"
    Event    string "build event (default: 'push') [$DRONE_BUILD_EVENT]"
    Finished int    "build finished (default: 0) [$DRONE_BUILD_FINISHED]"
    Link     string "build link [$DRONE_BUILD_LINK]"
    Number   int    "build number (default: 0) [$DRONE_BUILD_NUMBER]"
    Started  int    "build started (default: 0) [$DRONE_BUILD_STARTED]"
    Status   string "build status (default: 'success') [$DRONE_BUILD_STATUS]"
  }

  Commit {
    Author {
      Avatar string "git author avatar [$DRONE_COMMIT_AUTHOR_AVATAR]"
      Email  string "git author email [$DRONE_COMMIT_AUTHOR_EMAIL]"
      Name   string "git author name [$DRONE_COMMIT_AUTHOR]"
    }
    Branch  string "git commit branch (default: 'master') [$DRONE_COMMIT_BRANCH]"
    Link    string "git commit link [$DRONE_COMMIT_LINK]"
    Message string "git commit message [$DRONE_COMMIT_MESSAGE]"
    Ref     string "git commit ref (default: 'refs/heads/master') [$DRONE_COMMIT_REF]"
    Remote  string "git remote url [$DRONE_REMOTE_URL]"
    Sha     string "git commit sha [$DRONE_COMMIT_SHA]"
  }
}
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

Pull requests are welcome.

## Authors

* [@dim](https://github.com/dim)
* [@kakkoyun](https://github.com/kakkoyun)
* [@salimane](https://github.com/salimane)

## Copyright

See [LICENSE](LICENSE) document
