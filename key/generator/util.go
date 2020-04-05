package generator

import (
	"crypto/md5" // #nosec
	"fmt"
	"io"
)

// readerHasher generic md5 hash generater from io.Reader.
func readerHasher(readers ...io.Reader) (string, error) {
	// Use go1.14 new hashmap functions.
	h := md5.New() // #nosec

	for _, r := range readers {
		if _, err := io.Copy(h, r); err != nil {
			return "", fmt.Errorf("write reader as hash, %w", err)
		}
	}

	return fmt.Sprintf("%x", h.Sum(nil)), nil
}
