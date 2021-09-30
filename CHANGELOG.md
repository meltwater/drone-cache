# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## Unreleased

### Added

- Nothing.

### Changed

- Nothing.

### Removed

- Nothing.

### Deprecated

- Nothing.

## [1.2.1] - 2021-09-30

### Added

- [#183](https://github.com/meltwater/drone-cache/pull/183) set goarch for arm64 goreleaser

## [1.2.0] - 2021-09-29

**Warning** arm64 docker images are broken in this release, please use to 1.2.1

### Added

- [#146](https://github.com/meltwater/drone-cache/issues/146) Provide an arm image
  - Multiple PRs
- [#99](https://github.com/meltwater/drone-cache/issues/99) Document building images and pushing locally for PR testing
- [#142](https://github.com/meltwater/drone-cache/issues/142) backend/s3: Add option to assume AWS IAM role
- [#102](https://github.com/meltwater/drone-cache/pull/102) Implement option to disable cache rebuild if it already exists in storage.
- [#86](https://github.com/meltwater/drone-cache/pull/86) Add backend operation timeout option that cancels request if they take longer than given duration. `BACKEND_OPERATION_TIMEOUT`, `backend.operation-timeot`. Default value is `3 minutes`.
- [#86](https://github.com/meltwater/drone-cache/pull/86) Customize the cache key in the path. Adds a new `remote_root` option to customize it. Defaults to `repo.name`.
  - Fixes [#97](https://github.com/meltwater/drone-cache/issues/97).
  [#89](https://github.com/meltwater/drone-cache/pull/89) Add Azure Storage Backend.
  [#84](https://github.com/meltwater/drone-cache/pull/84) Adds compression level option.
  [#77](https://github.com/meltwater/drone-cache/pull/77) Adds a new hidden CLI flag to be used for tests.
  [#73](https://github.com/meltwater/drone-cache/pull/73) Add Google Cloud storage support
  [#68](https://github.com/meltwater/drone-cache/pull/68) Introduces new storage backend, sFTP.

### Changed

- [#138](https://github.com/meltwater/drone-cache/pull/138) backend/gcs: Fix GCS to pass credentials correctly when `GCS_ENDPOINT` is not set.
- [#135](https://github.com/meltwater/drone-cache/issues/135) backend/gcs: Fixed parsing of GCS JSON key.
- [#151](https://github.com/meltwater/drone-cache/issues/151) backend/s3: Fix assume role parameter passing
- [#164](https://github.com/meltwater/drone-cache/issues/164) tests: lock azurite image to 3.10.0
- [#133](https://github.com/meltwater/drone-cache/pull/133) backend/s3: Fixed Anonymous Credentials Error on public buckets. 
  - Fixes [#132](https://github.com/meltwater/drone-cache/issues/132)

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

## [1.1.0] - 2020-06-11

### Changed

- Docker login on CI
- Fix branch matchers
- Fix github actions
- Fix snapshot tag matcher
- [#121](https://github.com/meltwater/drone-cache/pull/121) Fix tests
- Generated missing embedded piece
- [#112](https://github.com/meltwater/drone-cache/pull/112) Improve documentation and development tooling
- [#125](https://github.com/meltwater/drone-cache/pull/125) Merge release v1.1
- [#124](https://github.com/meltwater/drone-cache/pull/124) Push images for snapshots
- Remove branch trigger
- Remove draft from releases prevents publishing
- Remove snapshot flag from goreleaser
- [#117](https://github.com/meltwater/drone-cache/pull/117) Removing leading newline in code block
- Revert snapshot simplifications
- Simplify snapshot releaser config
- [#111](https://github.com/meltwater/drone-cache/pull/111) Update docs for CLI args with new override flag
- [#126](https://github.com/meltwater/drone-cache/pull/126) Use bingo for tool dependencies
- [#127](https://github.com/meltwater/drone-cache/pull/127) Use latest release candidate in CI
- User docker token

### Removed

- Nothing.

### Deprecated

- Nothing.

## [1.0.4] - 2019-06-14

### Added

- [#42](https://github.com/meltwater/drone-cache/pull/42) Add symlinks to archive

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
