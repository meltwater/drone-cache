package autodetect

type nodePreparer struct{}

func newNodePreparer() *nodePreparer {
	return &nodePreparer{}
}

func (*nodePreparer) PrepareRepo() (string, error) {
	return "node_modules", nil
}
