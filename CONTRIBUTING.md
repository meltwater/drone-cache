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
1. Ensure any install or build dependencies are removed before the end of the layer when doing a
   build.
2. Please ensure the [README](README.md) and [DOCS](./DOCS.md) are up-to-date with details of changes to the command-line interface,
    this includes new environment variables, exposed ports, useful file locations and container parameters.
3. **PLEASE ENSURE YOU DO NOT INTRODUCE BREAKING CHANGES.**
4. **PLEASE ENSURE BUG FIXES AND NEW FEATURES INCLUDE TESTS.**
5. You may merge the Pull Request in once you have the sign-off of one other maintainer/code owner,
   or if you do not have permission to do that, you may request the second reviewer to merge it for you.

## Pull Request Checklist

- [x] Read the **CONTRIBUTING** document. (It's checked since you are already here.)
- [ ] Read the **CODE OF CONDUCT** document.
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
1. Increase the version numbers in any examples files and the README.md to the new version that this
   release would represent. The versioning scheme we use is [SemVer](http://semver.org/) for versioning. For the versions available, see the [tags on this repository](https://github.com/meltwater/drone-cache/tags).

2. Ensure [CHANGELOG](CHANGELOG.md) is up-to-date with new version changes.
3. Update version references.
4. Create a tag on master. Any changes on master will trigger a release with given tag and `latest tag.

    ```console
    $ git tag -am 'vX.X.X'
    > ...
    $ git push --tags
    > ...
    ```

> **Keep in mind that users usually use the `latest` tagged images in their pipeline, please make sure you do not interfere with their working workflow.**
