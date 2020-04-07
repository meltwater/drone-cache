package cache

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/meltwater/drone-cache/archive"
	"github.com/meltwater/drone-cache/internal"
	"github.com/meltwater/drone-cache/key"
	"github.com/meltwater/drone-cache/storage"

	"github.com/dustin/go-humanize"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
)

type rebuilder struct {
	logger log.Logger

	a  archive.Archive
	s  storage.Storage
	g  key.Generator
	fg key.Generator

	namespace string
	override  bool
}

// NewRebuilder TODO
func NewRebuilder(logger log.Logger, s storage.Storage, a archive.Archive, g key.Generator, fg key.Generator, namespace string, override bool) Rebuilder { //nolint:lll
	return rebuilder{logger, a, s, g, fg, namespace, override}
}

// Rebuild TODO
func (r rebuilder) Rebuild(srcs []string) error {
	level.Info(r.logger).Log("msg", "rebuilding cache")

	now := time.Now()

	key, err := r.generateKey()
	if err != nil {
		return fmt.Errorf("generate key, %w", err)
	}

	var (
		wg        sync.WaitGroup
		errs      = &internal.MultiError{}
		namespace = filepath.ToSlash(filepath.Clean(r.namespace))
	)

	for _, src := range srcs {
		if _, err := os.Lstat(src); err != nil {
			return fmt.Errorf("source <%s>, make sure file or directory exists and readable, %w", src, err)
		}

		dst := filepath.Join(namespace, key, src)

		// If no override is set and object already exists in storage, skip it.
		if !r.override {
			exists, err := r.s.Exists(dst)
			if err != nil {
				return fmt.Errorf("destination <%s> existence check, %w", dst, err)
			}

			if exists {
				continue
			}
		}

		level.Info(r.logger).Log("msg", "rebuilding cache for directory", "local", src, "remote", dst)

		wg.Add(1) //nolint:gomnd

		go func(dst, src string) {
			defer wg.Done()

			if err := r.rebuild(src, dst); err != nil {
				errs.Add(fmt.Errorf("upload from <%s> to <%s>, %w", src, dst, err))
			}
		}(dst, src)
	}

	wg.Wait()

	if errs.Err() != nil {
		return fmt.Errorf("rebuild failed, %w", errs)
	}

	level.Info(r.logger).Log("msg", "cache built", "took", time.Since(now))

	return nil
}

// rebuild pushes the archived file to the cache.
func (r rebuilder) rebuild(src, dst string) (err error) {
	src, err = filepath.Abs(filepath.Clean(src))
	if err != nil {
		return fmt.Errorf("clean source path, %w", err)
	}

	pr, pw := io.Pipe()
	defer internal.CloseWithErrCapturef(&err, pr, "rebuild, pr close <%s>", src)

	var written int64

	go func(wrt *int64) {
		defer internal.CloseWithErrLogf(r.logger, pw, "pw close defer")

		level.Info(r.logger).Log("msg", "archiving directory", "src", src)

		written, err := r.a.Create([]string{src}, pw)
		if err != nil {
			if err := pw.CloseWithError(fmt.Errorf("archive write, pipe writer failed, %w", err)); err != nil {
				level.Error(r.logger).Log("msg", "pw close", "err", err)
			}
		}

		*wrt += written
	}(&written)

	level.Info(r.logger).Log("msg", "uploading archived directory", "local", src, "remote", dst)

	sw := &statWriter{}
	tr := io.TeeReader(pr, sw)

	if err := r.s.Put(dst, tr); err != nil {
		err = fmt.Errorf("upload file, pipe reader failed, %w", err)
		if err := pr.CloseWithError(err); err != nil {
			level.Error(r.logger).Log("msg", "pr close", "err", err)
		}

		return err
	}

	level.Debug(r.logger).Log(
		"msg", "archive created",
		"local", src,
		"remote", dst,
		"archived bytes", humanize.Bytes(uint64(sw.written)),
		"read bytes", humanize.Bytes(uint64(written)),
		"ratio", fmt.Sprintf("%%%0.2f", float64(sw.written)/float64(written)*100.0), //nolint:gomnd
	)

	return nil
}

// Helpers

func (r rebuilder) generateKey(parts ...string) (string, error) {
	key, err := r.g.Generate(parts...)
	if err == nil {
		return key, nil
	}

	if r.fg != nil {
		level.Error(r.logger).Log("msg", "falling back to fallback key generator", "err", err)

		key, err = r.fg.Generate(parts...)
		if err == nil {
			return key, nil
		}
	}

	return "", err
}
