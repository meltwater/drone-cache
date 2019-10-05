# Drone 0.8 Examples

## Drone Configuration examples

> `!!!` The example Yaml configurations in this file are using the legacy 0.8 syntax. If you are using Drone 1.0 or Drone Cloud please ensure you use the appropriate 1.0 syntax. [Learn more here](https://docs.drone.io/config/pipeline/migrating/#plugins).

The following is a sample configuration in your .drone.yml file:

### Simple

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

### Simple (Filesystem/Volume)

NOTE: This will only be effective if your pipeline runs on the same agent each time (for
example, if you are running drone in single-machine mode).

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

### With custom cache key template

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
