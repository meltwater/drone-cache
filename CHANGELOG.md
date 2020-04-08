# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## Unreleased

### Added

- [#86](https://github.com/meltwater/drone-cache/pull/86) Add backend operation timeout option that cancels request if they take longer than given duration. `BACKEND_OPERATION_TIMEOUT`, `backend.operation-timeot`. Default value is `3 minutes`.
- [#86](https://github.com/meltwater/drone-cache/pull/86) Customize the cache key in the path. Adds a new `remote_root` option to customize it. Defaults to `repo.name`.
  - Fixes [#97](https://github.com/meltwater/drone-cache/issues/97).
[#89](https://github.com/meltwater/drone-cache/pull/89) Add Azure Storage Backend.
[#84](https://github.com/meltwater/drone-cache/pull/84) Adds compression level option.
[#77](https://github.com/meltwater/drone-cache/pull/77) Adds a new hidden CLI flag to be used for tests.
[#73](https://github.com/meltwater/drone-cache/pull/73) Add Google Cloud storage support
[#68](https://github.com/meltwater/drone-cache/pull/68) Introduces new storage backend, sFTP.

### Changed

- [#86](https://github.com/meltwater/drone-cache/pull/86) Support multipart uploads.
  - Fixes [#55](https://github.com/meltwater/drone-cache/issues/55).
  - Fixes [#88](https://github.com/meltwater/drone-cache/issues/88).
- [#86](https://github.com/meltwater/drone-cache/pull/86) Fixed rebuild errors when a nested path given in mount.
  - Fixes [#87](https://github.com/meltwater/drone-cache/issues/87).

### Removed

- Nothing.

### Deprecated

- Nothing.

## [1.0.4] - 2019-06-14

### Added

- Add symlinks to archive

### Changed

- Nothing

### Removed

- Nothing.

### Deprecated

- Nothing.

## [1.0.3] - 2019-06-11

### Added

- Add missing stage
- Add missing labels
- Add release latest
- Add snapshot stage

### Changed

f20a2ea Rename DRONE_REPO_OWNER to DRONE_REPO_NAMESPACE

### Removed

- Nothing.

### Deprecated

- Nothing.

## [1.0.2] - 2019-05-17

### Added

- Improvements to build pipeline

### Changed

9532da6 Clean and organize TODOs

### Removed

- Nothing.

### Deprecated

- Nothing.

## [1.0.1] - 2019-05-15

### Added

- Add blogpost link to it
- Add cache-key parameter to README example
- Add slack message image

### Changed

- Do not try to rebuild cache for the paths do not exist
- Fix drone release
- Fix image name in README
- Fix link to examples in README
- Fix parameter naming issue in examples
- Fix pure Docker example in README
- Some README improvements
- goreleaser releases Docker

### Removed

- Nothing.

### Deprecated

- Nothing.

## [1.0.0] - 2019-04-05

### Added

- Add new drone logo
- goreleaser releases Docker

### Changed

- Nothing

### Removed

- Nothing.

### Deprecated

- Nothing.

## [1.0.0-rc4] - 2019-04-05

### Added

- Add new drone logo
- goreleaser releases Docker

### Changed

- Nothing

### Removed

- Nothing.

### Deprecated

- Nothing.

## [1.0.0-rc3] - 2019-03-19

### Added

- Add docs for drone plugin index
- Add how it works
- Fix minor command issue
- Integrate with Drone.io
- Trigger on a tag
- Use scratch as base image

### Changed

ba005b6 Improve README

### Removed

- Nothing.

### Deprecated

- Nothing.

## [1.0.0-rc2] - 2019-03-05

### Added

- Fix behaviour with empty credentials
- Print out more information when debug enabled

### Changed

- Improve Documentation

### Removed

- Nothing.

### Deprecated

- Nothing.

## [1.0.0-rc1] - 2019-02-26

### Added

- Add additional information for cache keys
- Add annotations for cache metadata fields
- Add cache key template helper functions (checksum, epoch, arch, os)
- Add github codeowners
- Volume/Filesystem Cache (#15)

### Changed

- Enable more linters and fix discovered issues (#14)
- Update documentation (#16)

### Removed

- Nothing.

### Deprecated

- Nothing.

## [0.9.0] - 2019-02-15

### Added

- CircleCI like template cache keys
- Short names to CLI flags
- Gzip support
- integration tests

### Changed

- Make Restore/Rebuild flags mutually exclusive

### Removed

- Ability to read environment variables from a file removed.
- Plugin no longer depends on github.com/joho/godotenv. `env-file` flag is no longer available.
- Plugin no longer depends on github.com/sirupsen/logrus.

### Deprecated

- Nothing.
