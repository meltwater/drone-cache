package backend

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/meltwater/drone-cache/cache"
)

// filesystem is an file system implementation of the Backend
type filesystem struct {
	cacheRoot string
}

// newFileSystem returns a new file system Backend implementation
func newFileSystem(cacheRoot string) cache.Backend {
	return &filesystem{cacheRoot: cacheRoot}
}

// Get returns an io.Reader for reading the contents of the file
func (c *filesystem) Get(p string) (io.ReadCloser, error) {
	absPath, err := filepath.Abs(filepath.Clean(filepath.Join(c.cacheRoot, p)))
	if err != nil {
		return nil, fmt.Errorf("get the object %w", err)
	}

	return os.Open(absPath)
}

// Put uploads the contents of the io.ReadSeeker
func (c *filesystem) Put(p string, src io.ReadSeeker) error {
	absPath, err := filepath.Abs(filepath.Clean(filepath.Join(c.cacheRoot, p)))
	if err != nil {
		return fmt.Errorf("build path %w", err)
	}

	dir := filepath.Dir(absPath)
	if err := os.MkdirAll(dir, os.FileMode(0755)); err != nil { //nolint:mnd 755 is not a magic number
		return fmt.Errorf("create directory <%s> %w", dir, err)
	}

	dst, err := os.Create(absPath)
	if err != nil {
		return fmt.Errorf("create cache file <%s> %w", absPath, err)
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		return fmt.Errorf("write read seeker as file %w", err)
	}

	return nil
}
