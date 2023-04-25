package autodetect

import "path/filepath"

type nodePreparer struct{}

func newNodePreparer() *nodePreparer {
	return &nodePreparer{}
}

func (*nodePreparer) PrepareRepo(dir string) (string, error) {
	return filepath.Join(dir, "node_modules"), nil
}
