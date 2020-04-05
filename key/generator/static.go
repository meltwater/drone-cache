package generator

import "path/filepath"

type Static struct {
	defaultParts []string
}

// NewStatic TODO
func NewStatic(defaultParts ...string) *Static {
	return &Static{defaultParts: defaultParts}
}

// Generate generates key from given parts or templates as parameter.
func (s *Static) Generate(parts ...string) (string, error) {
	return filepath.Join(append(parts, s.defaultParts...)...), nil
}

// Check checks if generator functional.
func (s *Static) Check() error { return nil }
