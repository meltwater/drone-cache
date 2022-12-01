package auto_detect

type RepoPreparer interface {
	// PrepareRepo change local files to a state where cache intelligence options can be performed
	PrepareRepo() (string, error)
}
