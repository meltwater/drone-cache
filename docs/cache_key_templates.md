# Cache Key Templates

Cache key template syntax is very basic. You just need to provide a string. In that string you can use variables by prefixing them with a `.` in `{{ }}` construct, from provided metadata object (see below).

Also following helper functions provided for your use:

* `checksum`: Provides md5 hash of a file for given path
* `hashFiles`: Provides SHA256 hash after SHA256 hashing each single file
* `epoch`: Provides Unix epoch
* `arch`: Provides Architecture of running system
* `os`: Provides Operation system of running system

For further information about this syntax please see [official docs](https://golang.org/pkg/text/template/) from Go standard library.

## Template Examples

`"{{ .Repo.Name }}-{{ .Commit.Branch }}-{{ checksum "go.mod" }}-yadayadayada"`

`"{{ .Repo.Name }}_{{ checksum "go.mod" }}_{{ checksum "go.sum" }}_{{ arch }}_{{ os }}"`

`"{{ .Repo.Name }}_{{ hashFiles "go.mod" "go.sum" }}_{{ arch }}_{{ os }}"`

`"{{ .Repo.Name }}_{{ hashFiles "go.*" }}_{{ arch }}_{{ os }}"`

## Metadata

Following metadata object is available and pre-populated with current build information for you to use in cache key templates.

```go
{
  Repo {
    Avatar      string "repository avatar [$DRONE_REPO_AVATAR]"
    Branch      string "repository default branch [$DRONE_REPO_BRANCH]"
    Link        string "repository link [$DRONE_REPO_LINK]"
    Name        string "repository name [$DRONE_REPO_NAMESPACE]"
    Namespace   string "repository namespace [$DRONE_REPO_NAMESPACE]"
    Owner       string "repository owner [$DRONE_REPO_OWNER]"
    Private     bool   "repository is private [$DRONE_REPO_PRIVATE]"
    Trusted     bool   "repository is trusted [$DRONE_REPO_TRUSTED]"
  }

  Build {
    Created  int    "build created (default: 0) [$DRONE_BUILD_CREATED]"
    Deploy   string "build deployment target [$DRONE_DEPLOY_TO]"
    Event    string "build event (default: 'push') [$DRONE_BUILD_EVENT]"
    Finished int    "build finished (default: 0) [$DRONE_BUILD_FINISHED]"
    Link     string "build link [$DRONE_BUILD_LINK]"
    Number   int    "build number (default: 0) [$DRONE_BUILD_NUMBER]"
    Started  int    "build started (default: 0) [$DRONE_BUILD_STARTED]"
    Status   string "build status (default: 'success') [$DRONE_BUILD_STATUS]"
  }

  Commit {
    Author {
      Avatar string "git author avatar [$DRONE_COMMIT_AUTHOR_AVATAR]"
      Email  string "git author email [$DRONE_COMMIT_AUTHOR_EMAIL]"
      Name   string "git author name [$DRONE_COMMIT_AUTHOR]"
    }
    Branch  string "git commit branch (default: 'master') [$DRONE_COMMIT_BRANCH]"
    Link    string "git commit link [$DRONE_COMMIT_LINK]"
    Message string "git commit message [$DRONE_COMMIT_MESSAGE]"
    Ref     string "git commit ref (default: 'refs/heads/master') [$DRONE_COMMIT_REF]"
    Remote  string "git remote url [$DRONE_REMOTE_URL]"
    Sha     string "git commit sha [$DRONE_COMMIT_SHA]"
  }
}
```
