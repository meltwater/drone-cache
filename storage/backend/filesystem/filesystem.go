package filesystem

import (
	"context"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/meltwater/drone-cache/internal"
)

const defaultFileMode = 0755

// Backend is an file system implementation of the Backend.
type Backend struct {
	logger log.Logger

	cacheRoot string
}

// New creates a Backend backend.
func New(l log.Logger, c Config) (*Backend, error) {
	if strings.TrimRight(path.Clean(c.CacheRoot), "/") == "" {
		return nil, fmt.Errorf("empty or root path given, <%s> as cache root", c.CacheRoot)
	}

	level.Debug(l).Log("msg", "Filesystem backend", "config", fmt.Sprintf("%#v", c))

	//nolint: TODO(kakkoyun): Should it be created?
	if _, err := os.Stat(c.CacheRoot); err != nil {
		return nil, fmt.Errorf("make sure volume is mounted, <%s> as cache root, %w", c.CacheRoot, err)
	}

	return &Backend{logger: l, cacheRoot: c.CacheRoot}, nil
}

// Get writes downloaded content to the given writer.
func (b *Backend) Get(ctx context.Context, p string, w io.Writer) error {
	path, err := filepath.Abs(filepath.Clean(filepath.Join(b.cacheRoot, p)))
	if err != nil {
		return fmt.Errorf("absolute path, %w", err)
	}

	errCh := make(chan error)

	go func() {
		defer close(errCh)

		rc, err := os.Open(path)
		if err != nil {
			errCh <- fmt.Errorf("get the object, %w", err)
			return
		}

		defer internal.CloseWithErrLogf(b.logger, rc, "response body, close defer")

		_, err = io.Copy(w, rc)
		if err != nil {
			errCh <- fmt.Errorf("copy the object, %w", err)
			return
		}
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}

// Put uploads contents of the given reader.
func (b *Backend) Put(ctx context.Context, p string, r io.Reader) error {
	path, err := filepath.Abs(filepath.Clean(filepath.Join(b.cacheRoot, p)))
	if err != nil {
		return fmt.Errorf("build path, %w", err)
	}

	errCh := make(chan error)

	go func() {
		defer close(errCh)

		dir := filepath.Dir(path)
		if err := os.MkdirAll(dir, os.FileMode(defaultFileMode)); err != nil {
			errCh <- fmt.Errorf("create directory, %w", err)
		}

		w, err := os.Create(path)
		if err != nil {
			errCh <- fmt.Errorf("create cache file, %w", err)
			return
		}

		defer internal.CloseWithErrLogf(b.logger, w, "file writer, close defer")

		if _, err := io.Copy(w, r); err != nil {
			errCh <- fmt.Errorf("write contents of reader to a file, %w", err)
		}

		if err := w.Close(); err != nil {
			errCh <- fmt.Errorf("close the object, %w", err)
		}
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}

// Exists checks if object already exists.
func (b *Backend) Exists(ctx context.Context, p string) (bool, error) {
	path, err := filepath.Abs(filepath.Clean(filepath.Join(b.cacheRoot, p)))
	if err != nil {
		return false, fmt.Errorf("absolute path, %w", err)
	}

	_, err = os.Stat(path)
	if err != nil && !os.IsNotExist(err) {
		return false, fmt.Errorf("check the object exists, %w", err)
	}
	return err == nil, nil
}
