# TODO

## Maintenance

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

- [ ] Remove unused flags, simplify
- [ ] Make flags mutually exclusive, throw an error
- [ ] Add more build information
  - (using drone-start-pluging https://github.com/drone/drone-plugin-starter/)
  - Introduce plugin/definitions
- [ ] Add short names for Flags
  - (you can set alternate (or short) names for flags by providing a comma-delimited list for the Name.)
- [ ] Introduce compress and extract methods in Cache package
- [ ] Refactor tests
- [ ] Clean up TODOs

## Features

- [ ] Add CircleCI like go template cache keys
- [ ] Make sure cache fallbacks to master for default branched based cache

## In future

- [ ] Add unit tests
- [ ] Consider changing command-line framework
- [ ] Make object storage layer pluggable, introduces several providers
- [ ] Add documentation and examples, using go docs
  - ! (there is **no canonical way** to generate static docs and go doc requires an accessible github repo)
- [ ] Copyright
- [ ] Add reference to original author
- [ ] GitHub pages for documentation (like Distillery), MkDocs
- [ ] Provide Code of conduct
- [ ] Provide Contributors
- [ ] Provide PR template
- [ ] Provide issue template
- [ ] Open Source :tada:
