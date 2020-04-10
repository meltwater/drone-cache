# Drone 1.0 Examples

## Drone Configuration examples

The following is a sample configuration in your .drone.yml file:

### Simple

```yaml
kind: pipeline
name: default

steps:
  - name: restore-cache
    image: meltwater/drone-cache:dev
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
    image: golang:1.14.2-alpine3.11
    pull: true
    commands:
      - apk add --update make git
      - make drone-cache

  - name: rebuild-cache
    image: meltwater/drone-cache:dev
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

### Simple (Filesystem/Volume)

NOTE: This will only be effective if your pipeline runs on the same agent each time (for example, if you are running the drone in single-machine mode).

```yaml
kind: pipeline
name: default

steps:
  - name: restore-cache-with-filesystem
    image: meltwater/drone-cache:dev
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
    image: golang:1.14.2-alpine3.11
    pull: true
    commands:
      - apk add --update make git
      - make drone-cache

  - name: rebuild-cache-with-filesystem
    image: meltwater/drone-cache:dev
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
    host:
      path: /var/lib/cache
```

### With custom cache key template

See [cache key templates](../cache_key_templates.md#cache-key-templates) section for further information and to learn about syntax.

```yaml
kind: pipeline
name: default

steps:
  - name: restore-cache-with-key
    image: meltwater/drone-cache:dev
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
    image: golang:1.14.2-alpine3.11
    pull: true
    commands:
      - apk add --update make git
      - make drone-cache

  - name: rebuild-cache-with-key
    image: meltwater/drone-cache:dev
    pull: true
    environment:
      AWS_ACCESS_KEY_ID:
        from_secret: aws_access_key_id
      AWS_SECRET_ACCESS_KEY:
        from_secret: aws_secret_access_key
    settings:
      rebuild: true
      cache_key: '{{ .Repo.Name }}_{{ checksum "go.mod" }}_{{ checksum "go.sum" }}_{{ arch }}_{{ os }}'
      bucket: drone-cache-bucket
      region: eu-west-1
      mount:
        - 'vendor'
```

### With gzip compression

```yaml
kind: pipeline
name: default

steps:
  - name: restore-cache-with-gzip
    image: meltwater/drone-cache:dev
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
    image: golang:1.14.2-alpine3.11
    pull: true
    commands:
      - apk add --update make git
      - make drone-cache

  - name: rebuild-cache-with-gzip
    image: meltwater/drone-cache:dev
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

### Debug

```yaml
kind: pipeline
name: default

steps:
  - name: restore-cache-debug
    image: meltwater/drone-cache:dev
    settings:
      pull: true
      restore: true
      debug: true

  - name: build
    image: golang:1.14.2-alpine3.11
    pull: true
    commands:
      - apk add --update make git
      - make drone-cache

  - name: restore-cache-debug
    image: meltwater/drone-cache:dev
    settings:
      pull: true
      rebuild: true
      debug: true
```
