# TODO

### Maintenance

* [x] Add Drone badge
* [x] Merge DOCS.md and README.md
* [x] Use latest Go
* [x] Migrate to Go modules
* [x] Add UPX for binary compression

* [x] Add scaffold tests
* [x] Improve Drone build pipeline (add go static analyzers, test)

* [ ] Add integration tests (docker-compose, minio)
* [ ] Add unit tests

* [ ] Introduce compress and extract methods in Cache package
* [ ] Add more build information (using drone-start-pluging https://github.com/drone/drone-plugin-starter/)
    * Introduce plugin/definitions 
* [ ] Remove logrus
* [ ] Add pkg/errors
* [ ] Add more useful log messages
* [ ] Make flags mutually exclusive, throw an error

### Features

* [ ] Add CircleCI like go template cache keys
* [ ] Make sure cache fallbacks to master for default branched based cache

### In future

* [ ] Consider changing command-line framework
* [ ] Make object storage layer pluggable

* [ ] Copyright
* [ ] Add reference to original author
* [ ] Add documentation and examples, using go docs 
    * ! (there is no canonical way to generate static docs and go doc requires an accessible github repo)
* [ ] GitHub pages for documentation (like Distillery), MkDocs
* [ ] Open Source :tada:
