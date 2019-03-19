# TODO

## For v0.9.0

- [x] Add Drone badge
- [x] Merge DOCS.md and README.md
- [x] Use latest Go
- [x] Migrate to Go modules
- [x] Add UPX for binary compression
- [x] Add scaffold tests
- [x] Improve Drone build pipeline (add go static analyzers, test)
- [x] Remove logrus
- [x] Add integration tests
  - [x] docker-compose
  - [x] minio
- [x] Add pkg/errors
- [x] Fix broken tests, use ENV VARS to configure target object storage
- [x] Add more useful log messages and debug logs, clear useless log messages
- [x] Refactor tests
- [x] Clean up TODOs
- [x] Remove unused flags, simplify
- [x] Make flags mutually exclusive, throw an error
- [x] Add more build information
- [x] Add short names for Flags (you can set alternate (or short) names for flags by providing a comma-delimited list for the Name.)
- [x] Add Goreleaser
- [x] Add CircleCI like go template cache keys
- [x] Tar/Gzip
- [x] Add usage examples to README
- [x] Add all possible environment variables to README
- [x] Rename Repo
- [x] TEST!
- [x] MERGE!
- [x] Docker from scratch
- [x] Gorelease Docker multiple arc
- [x] Improve static analyzers
- [x] Badges
  - [x] Drone Pluging badges
  - [x] https://microbadger.com/images/meltwater/drone-cache
  - [x] https://goreportcard.com/report/github.com/meltwater/drone-cache

## v1.0.0-rc1

- [x] Fix `gosec`
- [x] Add helper functions for cache keys (https://circleci.com/docs/2.0/caching/#using-keys-and-templates)
  - [x] https://golang.org/pkg/text/template/#example_Template
  - [x] checksum (https://golang.org/pkg/crypto/md5/#New)
  - [x] epoch (https://gobyexample.com/epoch)
  - [x] arch (https://golang.org/pkg/runtime/#pkg-constants)
  - ~[ ] .Environment (https://gobyexample.com/environment-variables)~
- [X] **Add volume/file storage**
  - [x] https://docs.drone.io/user-guide/pipeline/volumes/
  - [x] http://plugins.drone.io/drillster/drone-volume-cache/
  - [x] https://github.com/Drillster/drone-volume-cache/blob/master/cacher.sh
- [x] Checkout
  - [x] https://github.com/drone/drone-go
  - [x] https://github.com/drone-plugins/drone-s3
  - [x] https://github.com/drone-plugins/drone-cache
  - [x] New Drone Version compatibility

## Before v1.0.0

- [ ] Improve documentation
  - [x] Examples
  - [ ] Drone 1.0 examples
- [x] Inspiration reference
    - https://github.com/bsm/drone-s3-cache
    - https://github.com/Drillster/drone-volume-cache
- [ ] Send to https://github.com/drone/drone-plugin-index

## Road to Open Source

- [x] Add Copyright
- [x] Add credits for original author [@dim](https://github.com/dim)
- [x] Update LICENCE
- [x] Project artwork
- [x] Improve README
- [ ] Open Source :tada:
- [ ] Add public CI/CD (https://cloud.drone.io/)

## Future work

- [ ] TTL/Retention policy or Flush
- [ ] Add Google Cloud Storage Backend
- [ ] Add SFTP Backend
- [ ] Add cache key fallback list
- [ ] Improve Makefile
- [ ] Add unit tests
