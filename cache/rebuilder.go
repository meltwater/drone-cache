package cache

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
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

	namespace      string
	override       bool
	gracefulDetect bool
}

// NewRebuilder creates a new cache.Rebuilder.
func NewRebuilder(logger log.Logger, s storage.Storage, a archive.Archive, g key.Generator, fg key.Generator, namespace string, override bool, gracefulDetect bool) Rebuilder { // nolint:lll
	return rebuilder{logger, a, s, g, fg, namespace, override, gracefulDetect}
}

// Rebuild rebuilds cache from the files provided with given paths.
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
			if !r.gracefulDetect {
				return fmt.Errorf("source <%s>, make sure file or directory exists and readable, %w", src, err)
			}
			level.Warn(r.logger).Log("msg", fmt.Sprintf("source directory %s does not exist", src), "err", fmt.Errorf("source <%s>, make sure file or directory exists and readable, %w", src, err))
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

		level.Info(r.logger).Log("msg", "rebuilding cache for directory", "local", src)
		level.Debug(r.logger).Log("msg", "rebuilding cache for directory", "remote", dst)

		wg.Add(1)

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
	isRelativePath := strings.HasPrefix(src, "./")
	level.Debug(r.logger).Log("msg", "rebuild", "src", src, "relativePath", isRelativePath) //nolint: errcheck
	src = filepath.Clean(src)
	if !isRelativePath {
		src, err = filepath.Abs(src)
		if err != nil {
			return fmt.Errorf("clean source path, %w", err)
		}
		level.Debug(r.logger).Log("msg", "src is adjusted", "src", src) //nolint: errcheck
	}

	pr, pw := io.Pipe()
	defer internal.CloseWithErrCapturef(&err, pr, "rebuild, pr close <%s>", src)

	var written int64

	go func(wrt *int64) {
		defer internal.CloseWithErrLogf(r.logger, pw, "pw close defer")

		level.Debug(r.logger).Log("msg", "caching paths", "src", src)

		localWritten, err := r.a.Create([]string{src}, pw, isRelativePath)
		if err != nil {
			if err := pw.CloseWithError(fmt.Errorf("archive write, pipe writer failed, %w", err)); err != nil {
				level.Error(r.logger).Log("msg", "pw close", "err", err)
			}
		}

		*wrt += localWritten
	}(&written)

	level.Debug(r.logger).Log("msg", "uploading archived directory", "local", src, "remote", dst)

	sw := &statWriter{}
	tr := io.TeeReader(pr, sw)

	if err := r.s.Put(dst, tr); err != nil {
		err = fmt.Errorf("upload file, pipe reader failed, %w", err)
		if err := pr.CloseWithError(err); err != nil {
			level.Error(r.logger).Log("msg", "pr close", "err", err)
		}

		return err
	}

	level.Info(r.logger).Log("msg", "uploaded cache", "src", src, "total", humanize.Bytes(uint64(sw.written)), "comressed", humanize.Bytes(uint64(written)))

	level.Debug(r.logger).Log(
		"msg", "archive created",
		"local", src,
		"remote", dst,
		"archived bytes", humanize.Bytes(uint64(sw.written)),
		"read bytes", humanize.Bytes(uint64(written)),
		"ratio", fmt.Sprintf("%%%0.2f", float64(sw.written)/float64(written)*100.0), // nolint:gomnd
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
