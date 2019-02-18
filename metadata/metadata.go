package metadata

type (
	// Repo stores information about repository that is built
	Repo struct {
		Avatar  string
		Branch  string
		Link    string
		Name    string
		Owner   string
		Private bool
		Trusted bool
	}

	// Build stores information about current build
	Build struct {
		Created  int64
		Deploy   string
		Event    string
		Finished int64
		Link     string
		Number   int
		Started  int64
		Status   string
	}

	// Commit stores information about current commit
	Commit struct {
		Author  Author
		Branch  string
		Link    string
		Message string
		Ref     string
		Remote  string
		Sha     string
	}

	// Author stores information about current commit's author
	Author struct {
		Avatar string
		Email  string
		Name   string
	}

	Metadata struct {
		Build  Build
		Commit Commit
		Repo   Repo
	}
)
