
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
    pull: true
    settings:
      restore: true
      bucket: drone-cache-bucket
      settings:
      aws_access_key_id:
        from_secret: aws_access_key_id
      aws_secret_access_key:
        from_secret: aws_secret_access_key
      region: eu-west-1
      mount:
        - 'vendor'

  - name: build
    image: golang:1.11-alpine
    pull: true
    commands:
      - apk add --update make git
      - make drone-cache

  - name: rebuild-cache
    image: meltwater/drone-cache:dev
    pull: true
    settings:
      rebuild: true
      bucket: drone-cache-bucket
      aws_access_key_id:
        from_secret: aws_access_key_id
      aws_secret_access_key:
        from_secret: aws_secret_access_key
      region: eu-west-1
      mount:
        - 'vendor'
```

### Simple (Filesystem/Volume)

```yaml
kind: pipeline
name: default

steps:
  - name: restore-cache
    image: meltwater/drone-cache:dev
    pull: true
    settings:
      restore: true
      backend: "filesystem" # (default: s3)
      bucket: drone-cache-bucket
      settings:
      aws_access_key_id:
        from_secret: aws_access_key_id
      aws_secret_access_key:
        from_secret: aws_secret_access_key
      region: eu-west-1
      mount:
        - 'vendor'

  - name: build
    image: golang:1.11-alpine
    pull: true
    commands:
      - apk add --update make git
      - make drone-cache

  - name: rebuild-cache
    image: meltwater/drone-cache:dev
    pull: true
    settings:
      rebuild: true
      backend: "filesystem" # (default: s3)
      bucket: drone-cache-bucket
      aws_access_key_id:
        from_secret: aws_access_key_id
      aws_secret_access_key:
        from_secret: aws_secret_access_key
      region: eu-west-1
      mount:
        - 'vendor'
```

### With custom cache key prefix template

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

### With gzip compression

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

### Debug

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
