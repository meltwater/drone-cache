package autodetect

type goPreparer struct{}

func newGoPreparer() *goPreparer {
	return &goPreparer{}
}

func (*goPreparer) PrepareRepo(dir string) (string, error) {
	return ".go", nil
}
