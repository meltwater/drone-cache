---
date: 2019-03-19T00:00:00+00:00
title: Drone Cache
author: meltwater
tags: [ cache, amazon, aws, s3, azure, volume ]
logo: drone_cache.svg
repo: meltwater/drone-cache
image: meltwater/drone-cache
---

{{% alert error %}}
This plugin requires Volume configuration if you enable `filesystem` backend with configuration. This means your repository Trusted flag must be enabled. This should not be enabled in untrusted environments.
{{% /alert %}}

A Drone plugin for caching current workspace files between builds to reduce your build times. `drone-cache` is a small CLI program, written in Go without any external OS dependencies (such as tar, etc).

With `drone-cache`, you can provide your **own cache key templates**, specify **archive format** (tar, tar.gz, etc) and you can use **an S3 bucket, Azure Storage or a mounted volume** as storage for your cached files, even better you can implement **your own storage backend** to cover your use case.

**How does it work**

`drone-cache` stores mounted directories and files under a key at the specified backend (by default S3).

Use this plugin to cache data that makes your builds faster. In the case of a cache miss or zero cache restore it will fail silently in won't break your running pipeline.

The best example would be to use this with your package managers such as Mix, Bundler or Maven. After your initial download, you can build a cache and then you can restore that cache in your next build.

With restored dependencies from a cache, commands like `mix deps.get` will only need to download new dependencies, rather than re-download every package on each and every build.

# Using Cache Key Templates

Cache key template syntax is very basic. You just need to provide a string. In that string you can use variables by prefixing them with a `.` in `{{ }}` construct, from provided metadata object (see below).

Also following helper functions provided for your use:

* `checksum`: Provides md5 hash of a file for given path
* `epoch`: Provides Unix epoch
* `arch`: Provides Architecture of running system
* `os`: Provides Operation system of running system

For further information about this syntax please see [official docs](https://golang.org/pkg/text/template/) from Go standard library.

**Template Examples**

`"{{ .Repo.Name }}-{{ .Commit.Branch }}-{{ checksum "go.mod" }}-yadayadayada"`

`"{{ .Repo.Name }}_{{ checksum "go.mod" }}_{{ checksum "go.sum" }}_{{ arch }}_{{ os }}"`
*Metadata*

Following metadata object is available and pre-populated with current build information for you to use in cache key templates.

```go
{
  Repo {
    Avatar    string "repository avatar [$DRONE_REPO_AVATAR]"
    Branch    string "repository default branch [$DRONE_REPO_BRANCH]"
    Link      string "repository link [$DRONE_REPO_LINK]"
    Name      string "repository name [$DRONE_REPO_NAME]"
    Owner     string "repository owner [$DRONE_REPO_NAMESPACE]"
    Namespace string "repository namespace [$DRONE_REPO_NAMESPACE]"
    Private   bool   "repository is private [$DRONE_REPO_PRIVATE]"
    Trusted   bool   "repository is trusted [$DRONE_REPO_TRUSTED]"
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

# Examples

The following is a sample configuration in your .drone.yml file:

**Simple**

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
      bucket: drone-cache-bucket
      region: eu-west-1
      mount:
        - 'vendor'

  - name: build
    image: golang:1.18.4
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
      bucket: drone-cache-bucket
      region: eu-west-1
      mount:
        - 'vendor'
```

**Simple (Filesystem/Volume)**

```yaml
kind: pipeline
name: default

steps:
  - name: restore-cache-with-filesystem
    image: meltwater/drone-cache
    pull: true
    settings:
      backend: "filesystem"
      restore: true
      cache_key: "volume"
      archive_format: "gzip"
      # filesystem_cache_root: "/tmp/cache"
      mount:
        - 'vendor'
    volumes:
    - name: cache
      path: /tmp/cache

  - name: build
    image: golang:1.18.4
    pull: true
    commands:
      - apk add --update make git
      - make drone-cache

  - name: rebuild-cache-with-filesystem
    image: meltwater/drone-cache
    pull: true
    settings:
      backend: "filesystem"
      rebuild: true
      cache_key: "volume"
      archive_format: "gzip"
      # filesystem_cache_root: "/tmp/cache"
      mount:
        - 'vendor'
    volumes:
    - name: cache
      path: /tmp/cache

volumes:
  - name: cache
    temp: {}
```

**With custom cache key template**

See [cache key templates](#using-cache-key-templates) section for further information and to learn about syntax.

```yaml
kind: pipeline
name: default

steps:
  - name: restore-cache-with-key
    image: meltwater/drone-cache
    environment:
      AWS_ACCESS_KEY_ID:
        from_secret: aws_access_key_id
      AWS_SECRET_ACCESS_KEY:
        from_secret: aws_secret_access_key
    settings:
      pull: true
      restore: true
      cache_key: '{{ .Repo.Name }}_{{ checksum "go.mod" }}_{{ checksum "go.sum" }}_{{ arch }}_{{ os }}'
      bucket: drone-cache-bucket
      region: eu-west-1
      mount:
        - 'vendor'

  - name: build
    image: golang:1.18.4
    pull: true
    commands:
      - apk add --update make git
      - make drone-cache

  - name: rebuild-cache-with-key
    image: meltwater/drone-cache
    pull: true
    environment:
      AWS_ACCESS_KEY_ID:
        from_secret: aws_access_key_id
      AWS_SECRET_ACCESS_KEY:
        from_secret: aws_secret_access_key
    settings:
      rebuild: true
      override: false
      cache_key: '{{ .Repo.Name }}_{{ checksum "go.mod" }}_{{ checksum "go.sum" }}_{{ arch }}_{{ os }}'
      bucket: drone-cache-bucket
      region: eu-west-1
      mount:
        - 'vendor'
```

*With gzip compression*

```yaml
kind: pipeline
name: default

steps:
  - name: restore-cache-with-gzip
    image: meltwater/drone-cache
    pull: true
    environment:
      AWS_ACCESS_KEY_ID:
        from_secret: aws_access_key_id
      AWS_SECRET_ACCESS_KEY:
        from_secret: aws_secret_access_key
    settings:
      restore: true
      cache_key: "gzip"
      archive_format: "gzip"
      bucket: drone-cache-bucket
      region: eu-west-1
      mount:
        - 'vendor'

  - name: build
    image: golang:1.18.4
    pull: true
    commands:
      - apk add --update make git
      - make drone-cache

  - name: rebuild-cache-with-gzip
    image: meltwater/drone-cache
    pull: true
    environment:
      AWS_ACCESS_KEY_ID:
        from_secret: aws_access_key_id
      AWS_SECRET_ACCESS_KEY:
        from_secret: aws_secret_access_key
    settings:
      rebuild: true
      cache_key: "gzip"
      archive_format: "gzip"
      bucket: drone-cache-bucket
      region: eu-west-1
      mount:
        - 'vendor'
```

**Debug**

```yaml
kind: pipeline
name: default

steps:
  - name: restore-cache-debug
    image: meltwater/drone-cache
    settings:
      pull: true
      restore: true
      debug: true

  - name: build
    image: golang:1.18.4
    pull: true
    commands:
      - apk add --update make git
      - make drone-cache

  - name: restore-cache-debug
    image: meltwater/drone-cache
    settings:
      pull: true
      rebuild: true
      debug: true
```

# Parameter Reference

backend
: cache backend to use in plugin (`s3`, `filesystem`) (default: `s3`)

mount
: cache directories, an array of folders to cache

rebuild
: rebuild the cache directories

restore
: restore the cache directories

cache_key
: cache key to use for the cache directories

archive_format
: archive format to use to store the cache directories (`tar`, `gzip`, `zstd`) (default: `tar`)

override
: override already existing cache files (default: `true`)

debug
: enable debug

filesystem-cache-root
: local filesystem root directory for the filesystem cache (default: `/tmp/cache`)

endpoint
: endpoint for the s3 connection

access_key
: AWS access key

secret_key
: AWS secret key

bucket
: AWS bucket name

region
: AWS bucket region. (`us-east-1`, `eu-west-1`, ...)

account_name
: Azure Storage account name

account_key
: Azure Storage account key

container
: Azure Storage container

path-style
: use path style for bucket paths. (true for `minio`, false for `aws`)

acl
: upload files with acl (`private`, `public-read`, ...) (default: `private`)

encryption
: server-side encryption algorithm, defaults to `none`. (`AES256`, `aws:kms`)

skip_symlinks
: skip symbolic links in archive
