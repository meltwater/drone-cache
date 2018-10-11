# TODO

### Maintenance

* [x] Add Drone badge
* [x] Merge DOCS.md and README.md
* [x] Use latest Go
* [x] Migrate to Go modules
* [x] Add UPX for binary compression

* [ ] Add integration tests (docker-compose, minio)
* [ ] Add tests and examples

* [ ] Improve Drone build pipeline (add go static analyzers, test)

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
* [ ] Add documentation using go docs 
    * ! (there is no canonical way to generate static docs and go doc requires an accessible github repo)
* [ ] GitHub pages for documentation (like Distillery), MkDocs
* [ ] Open Source :tada:
