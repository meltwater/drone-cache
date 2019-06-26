<!--
**IMPORTANT: Please do not create a Pull Request without creating an issue first.**

*Any change needs to be discussed before proceeding. Failure to do so may result in the rejection of the pull request.*

Thank you for your pull request. Please provide a description above and review
the requirements below.

Bug fixes and new features should include tests.

Contributors guide: ./CONTRIBUTING.md
-->

Please read [CONTRIBUTING.md](CONTRIBUTING.md) for details on our code of conduct, and the process for submitting pull requests to us.

Fixes #

## Proposed Changes
<!-- Please briefly list the changes you made here. -->

-
-
-

### Description

<!-- Please explain the changes you made here. -->

### Checklist

<!-- _Please make sure to review and check all of these items:_ -->

<!-- Remove items that do not apply. For completed items, change [ ] to [x]. -->

- [ ] Read the **CONTRIBUTING** document.
- [ ] Add tests to cover my changes.
- [ ] Make sure your code follows the code style of this project
- [ ] Make sure CI and all other PR checks are green OR
    - [ ] Code compiles correctly/
    - [ ] Created tests which fail without the change (if possible)/
    - [ ] All new and existing tests passed.
- [ ] Extended and update the README (if necessary).
- [ ] Make sure [documentation](./DOCS.md) is up-to-date. Same file should also be updated in [plugin index](https://github.com/drone/drone-plugin-index/blob/master/content/meltwater/drone-cache/index.md) by sending a PR to that repository.

**PLEASE DO NOT INTRODUCE BREAKING CHANGES**

**Keep in mind that users usually use the `latest` tagged images in their pipeline, please make sure you do not interfere with their working workflow.**

- [ ] Version bump if you have to, update version in documentation (fix or feature that would cause existing functionality to change).
    - We use [SemVer](http://semver.org/) for versioning. For the versions available, see the [tags on this repository](https://github.com/meltwater/drone-cache/tags).
