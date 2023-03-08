package autodetect

type goPreparer struct{}

func newGoPreparer() *goPreparer {
	return &goPreparer{}
}

func (*goPreparer) PrepareRepo() (string, error) {
	return ".go", nil
}
