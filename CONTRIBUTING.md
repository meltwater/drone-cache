# Contributing

When contributing to this repository, please first discuss the change you wish to make via issue,
email, or any other method with the owners of this repository before making a change.

Please note we have a [code of conduct](CODE_OF_CONDUCT.md) that we ask you to follow in all your interactions with the project.

**IMPORTANT: Please do not create a Pull Request without creating an issue first.**

*Any change needs to be discussed before proceeding. Failure to do so may result in the rejection of the pull request.*

Thank you for your pull request. Please provide a description above and review
the requirements below.

## Pull Request Process

0. Check out [Pull Request Checklist](#pull-request-checklist), ensure you have fulfilled each step.
1. Check out guidelines below, the project tries to follow these, ensure you have fulfilled them as much as possible.
    * [Effective Go](https://golang.org/doc/effective_go.html)
    * [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
2. Ensure any install or build dependencies are removed before the end of the layer when doing a
   build.
3. Please ensure the [README](README.md) and [DOCS](./DOCS.md) are up-to-date with details of changes to the command-line interface,
    this includes new environment variables, exposed ports, used file locations, and container parameters.
4. **PLEASE ENSURE YOU DO NOT INTRODUCE BREAKING CHANGES.**
5. **PLEASE ENSURE BUG FIXES AND NEW FEATURES INCLUDE TESTS.**
6. You may merge the Pull Request in once you have the sign-off of one other maintainer/code owner,
   or if you do not have permission to do that, you may request the second reviewer to merge it for you.

## Pull Request Checklist

- [x] Read the **CONTRIBUTING** document. (It's checked since you are already here.)
- [ ] Read the [**CODE OF CONDUCT**](CODE_OF_CONDUCT.md) document.
- [ ] Add tests to cover changes.
- [ ] Ensure your code follows the code style of this project.
- [ ] Ensure CI and all other PR checks are green OR
    - [ ] Code compiles correctly.
    - [ ] Created tests which fail without the change (if possible).
    - [ ] All new and existing tests passed.
- [ ] Add your changes to `Unreleased` section of [CHANGELOG](CHANGELOG.md).
- [ ] Improve and update the [README](README.md) (if necessary).
- [ ] Ensure [documentation](./DOCS.md) is up-to-date. The same file will be updated in [plugin index](https://github.com/drone/drone-plugin-index/blob/master/content/meltwater/drone-cache/index.md) when your PR is accepted, so it will be available for end-users at http://plugins.drone.io.

## Release Process

*Only concerns maintainers/code owners*

0. **PLEASE DO NOT INTRODUCE BREAKING CHANGES**
1. Execute `make README.md`. This will update [usage](README.md#usage) section of [README.md](README.md) with latest CLI options
2. Increase the version numbers in any examples files and the README.md to the new version that this
   the release would represent. The versioning scheme we use is [SemVer](http://semver.org/) for versioning. For the versions available, see the [tags on this repository](https://github.com/meltwater/drone-cache/tags).

3. Ensure [CHANGELOG](CHANGELOG.md) is up-to-date with new version changes.
4. Update version references.
5. Create a tag on the master. Any changes on the master will trigger a release with the given tag and `latest tag.

    ```console
    $ git tag -am 'vX.X.X'
    > ...
    $ git push --tags
    > ...
    ```
6. Check whether all the generate artifacts in-place properly.
7. Update [plugin index](https://github.com/drone/drone-plugin-index/blob/master/content/meltwater/drone-cache/index.md) using [DOCS](./DOCS.md).

> **Keep in mind that users usually use the `latest` tagged images in their pipeline, please make sure you do not interfere with their working workflow.**

## Testing Locally

Want to test locally without opening a PR?  Follow the steps below to build a local image of drone-cache and run the Drone pipeline against it.

0. Make sure you have the [Drone CLI](https://docs.drone.io/cli/install/),  [Docker](https://docs.docker.com/get-docker/), and [GoReleaser](https://goreleaser.com/install/) installed locally.
1. Update the `image_templates` key in `drone-cache/.goreleaser-local.yml`  to reflect the name you'd like your image to have, then run `release --config=.goreleaser-local.yml --snapshot --skip-publish --rm-dist` to build the image.
2. Update the `image: drone-cache:MyTestTag` entries in the `local-pipeline` pipeline in the `.drone.yml` with the name of the image that you created (there are several of these).
3. Run the Drone pipeline locally with `drone exec --branch MyBranchName --pipeline local-pipeline`

