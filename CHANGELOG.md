# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## Unreleased

- [#141](https://github.com/meltwater/drone-cache/pull/141) archive/tar, archive/gzip:
  add absolute path mode: fix an issue #130 with drone where it fails to make extraction if the passed path is an absoulte path.
### Added

- [#223](https://github.com/meltwater/drone-cache/pull/223) Added implementation for AlibabaOSS for backend storage

### Changed

- Updated dependency `cloud.google.com/go/storage v1.24.0` -> `v1.26.0`
- Updated dependency `github.com/urfave/cli/v2 v2.11.1` -> `v2.14.1`
- Updated dependency `google.golang.org/api v0.88.0` -> `v0.94.0`
- Updated dependency `google.golang.org/protobuf v1.28.0 ` -> `v1.28.1`

### Removed

## [1.4.0] - 2022-09-01

### Added

- [#209](https://github.com/meltwater/drone-cache/pull/209) Added double star directory searching in mounts (e.g. `path/**/subdir`)
- [#198](https://github.com/meltwater/drone-cache/pull/198) Add `hashFiles` template function to generate the SHA256 hash of multiple files

### Changed

- Bumped base go version from Go 1.14 to 1.18
- Switched and updated to moved dependency `go-kit/kit@v0.9.0` -> `go-kit/log@v0.2.1`
- Updated dependency `Azure/azure-storage-blob-go@v0.8.0` -> `v0.15.0`
- Updated dependency `aws/aws-sdk-go@v1.37.29` -> `v1.44.55`
- Updated dependency `cloud.google.com/go/storage@v1.1.0` -> `v1.23.0`
- Updated dependency `google.golang.org/api@v0.9.0` -> `v0.87.0`
- Updated dependency `google/go-cmp@v0.4.0` -> `v0.5.8`
- Updated dependency `klauspost/compress@v1.13.5` -> `v1.15.8`
- Updated dependency `pkg/sftp@v1.10.1` -> `v1.13.5`
- Updated dependency `urface/cli/v2@v2.1.1` -> `v2.11.0`
- Updated linting from manual install to official `golangci/golangci-lint:v1.46.2` Docker image 
- Updated golang base image from `1.14-alpine` to `1.18.4` (debian); issues with alpine `>= 3.13` due to DroneCI Docker Engine version
- Updated test image `minio/minio:RELEASE.2020-11-06T23-17-07Z` to `RELEASE.2022-07-15T03-44-22Z`
- Updated test image `fsouza/fake-gcs-server:1.18.3` to `1.38.3`
- Updated test image `mcr.microsoft.com/azure-storage/azurite:3.10.0` to `3.18.0`
- Linting fixes for older Go version code style
- Added missing struct argument for Azure Blob URLs (`azblob.ClientProvidedKeyOptions{}`)
- Renamed test cases to comply with Azure API disallowing non-alphanumeric characters in storage requests
- Linting rules adjusted to omit undesirable linters (see `.golanci.yml`)

### Removed

- Pushing `DOCS.md` Drone Plugin documenation; Drone/Harness now pull READMEs from Plugin repos

### Deprecated

- Nothing.

## [1.3.0] - 2022-04-05

### Added

- [#197](https://github.com/meltwater/drone-cache/pull/197) Zstd support

### Changed

- [#191](https://github.com/meltwater/drone-cache/issues/191) Update examples to reference non-dev images

## [1.2.2] - 2021-10-01

- [#188](https://github.com/meltwater/drone-cache/pull/188) v1.2.0 breaks EC2 IAM role bucket access

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
