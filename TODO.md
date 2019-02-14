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

- [ ] Add usage examples to README
- [ ] Add all possible environment variables to README
- [ ] Rename Repo

- [] MERGE !

- [ ] Improve static analyzers
- [ ] Improve Makefile
- [ ] Docker from scratch

## Before v1.0.0

- [ ] Introduce mode: wrap_in_directory: true
- [ ] Add volume/file storage
    - [ ] https://docs.drone.io/user-guide/pipeline/volumes/
    - [ ] http://plugins.drone.io/drillster/drone-volume-cache/
    - [ ] https://github.com/Drillster/drone-volume-cache/blob/master/cacher.sh
- [ ] Add unit tests
- [ ] Badges
    - [ ] Drone Pluging badges
    - [ ] https://microbadger.com/images/meltwater/drone-s3-cache
    - [ ] [![Go Report Card](https://goreportcard.com/badge/github.com/meltwater/drone-cache)](https://goreportcard.com/report/github.com/meltwater/drone-cache)

## Road to Open Source

- [ ] Copyright
- [ ] Add reference to original author
- [ ] LICENCE
- [ ] Provide Code of conduct
- [ ] Provide Contributors
- [ ] Provide PR template
- [ ] Provide Issue template
- [ ] Open Source :tada:
