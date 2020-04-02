package generator

import (
	"fmt"
	"io"
	"strings"
)

// Hash TODO
type Hash struct {
	defaultParts []string
}

// NewHash TODO
func NewHash(defaultParts ...string) *Hash {
	return &Hash{defaultParts: defaultParts}
}

// Generate generates key from given parts or templates as parameter.
func (h *Hash) Generate(parts ...string) (string, error) {
	key, err := hash(append(parts, h.defaultParts...)...)
	if err != nil {
		return "", fmt.Errorf("generate hash key for mounted %w", err)
	}

	return key, nil
}

// Check checks if generator functional.
func (h *Hash) Check() error { return nil }

// hash generates a key based on given strings (ie. filename paths and branch).
func hash(parts ...string) (string, error) {
	readers := make([]io.Reader, len(parts))
	for i, p := range parts {
		readers[i] = strings.NewReader(p)
	}

	return readerHasher(readers...)
}
