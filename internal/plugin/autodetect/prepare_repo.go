package autodetect

type RepoPreparer interface {
	// PrepareRepo change local files to a state where cache intelligence options can be performed
	PrepareRepo(dir string) (string, error)
}
