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
- [x] MERGE !
- [x] Docker from scratch
- [x] Gorelease Docker multiple arc
- [x] Improve static analyzers
- [x] Badges
  - [x] Drone Pluging badges
  - [x] https://microbadger.com/images/meltwater/drone-cache
  - [x] https://goreportcard.com/report/github.com/meltwater/drone-cache

## v0.10.0

- [x] Fix `gosec`
- [x] Add helper functions for cache keys (https://circleci.com/docs/2.0/caching/#using-keys-and-templates)
  - [x] https://golang.org/pkg/text/template/#example_Template
  - [x] checksum (https://golang.org/pkg/crypto/md5/#New)
  - [x] epoch (https://gobyexample.com/epoch)
  - [x] arch (https://golang.org/pkg/runtime/#pkg-constants)
  - ~[ ] .Environment (https://gobyexample.com/environment-variables)~

## v0.11.0

- [ ] **Add volume/file storage**
  - [x] https://docs.drone.io/user-guide/pipeline/volumes/
  - [x] http://plugins.drone.io/drillster/drone-volume-cache/
  - [x] https://github.com/Drillster/drone-volume-cache/blob/master/cacher.sh
- [x] Checkout
  - [x] https://github.com/drone/drone-go
  - [x] https://github.com/drone-plugins/drone-s3
  - [x] https://github.com/drone-plugins/drone-cache
  - [x] New Drone Version compatibility
- [ ] Improve documentation
  - [x] Examples
  - [ ] Drone 1.0 examples

## Before v1.0.0

- [ ] Fix tmp directory create permissions for scratch/unprivileged user in container
- [ ] Improve Makefile
- [ ] Clean up TODOs
- [ ] Add unit tests

## Road to Open Source

- [ ] Add Copyright
- [ ] Add credits for original author [@dim](https://github.com/dim)
- [ ] Update LICENCE
- [ ] Open Source :tada:

## Future work

- [ ] TTL/Retention policy
