package generator

import (
	"fmt"
	hash2 "hash"
	"io"
	"strings"
)

// Hash implements a key generator that uses the specified hash algorithm.
type Hash struct {
	hasher       func() hash2.Hash
	defaultParts []string
}

// NewHash creates a new hash key generator.
func NewHash(hasher func() hash2.Hash, defaultParts ...string) *Hash {
	return &Hash{
		hasher:       hasher,
		defaultParts: defaultParts,
	}
}

// Generate generates key from given parts or templates as parameter.
func (h *Hash) Generate(parts ...string) (string, error) {
	parts = append(parts, h.defaultParts...)
	readers := make([]io.Reader, len(parts))

	for i, p := range parts {
		readers[i] = strings.NewReader(p)
	}

	key, err := readerHasher(h.hasher, readers...)
	if err != nil {
		return "", fmt.Errorf("generate hash key for mounted, %w", err)
	}

	return fmt.Sprintf("%x", key), nil
}

// Check checks if generator functional.
func (h *Hash) Check() error { return nil }
