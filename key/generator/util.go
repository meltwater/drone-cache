package generator

import (
	"fmt"
	hash2 "hash"
	"io"
)

func readerHasher(hasher func() hash2.Hash, readers ...io.Reader) ([]byte, error) {
	h := hasher()

	for _, r := range readers {
		if _, err := io.Copy(h, r); err != nil {
			return nil, fmt.Errorf("write reader as hash, %w", err)
		}
	}

	return h.Sum(nil), nil
}
