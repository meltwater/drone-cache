// Package cache provides functionality for cache storage
package cache

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
)

// Backend implements operations for caching files
type Backend interface {
	Get(string) (io.ReadCloser, error)
	Put(string, io.ReadSeeker) error
}

// Cache contains configuration for Cache functionality
type Cache struct {
	logger log.Logger

	b    Backend
	opts options
}

// New creates a new cache with given parameters
func New(logger log.Logger, b Backend, opts ...Option) Cache {
	options := options{
		archiveFmt:       DefaultArchiveFormat,
		compressionLevel: DefaultCompressionLevel,
	}

	for _, o := range opts {
		o.apply(&options)
	}

	return Cache{
		logger: log.With(logger, "component", "cache"),
		b:      b,
		opts:   options,
	}
}

// Push pushes the archived file to the cache
func (c Cache) Push(src, dst string) error {
	// 1. check if source is reachable
	src, err := filepath.Abs(filepath.Clean(src))
	if err != nil {
		return fmt.Errorf("read source directory %w", err)
	}

	level.Info(c.logger).Log("msg", "archiving directory", "src", src)

	// 2. create a temporary file for the archive
	if err := os.MkdirAll("/tmp", os.FileMode(0755)); err != nil {
		return fmt.Errorf("create tmp directory %w", err)
	}

	dir, err := ioutil.TempDir("", "")
	if err != nil {
		return fmt.Errorf("create tmp folder for archive %w", err)
	}

	archivePath := filepath.Join(dir, "archive.tar")

	file, err := os.Create(archivePath)
	if err != nil {
		return fmt.Errorf("create tarball file <%s> %w", archivePath, err)
	}

	tw, twCloser, err := archiveWriter(file, c.opts.archiveFmt, c.opts.compressionLevel)
	if err != nil {
		return fmt.Errorf("initialize archive writer %w", err)
	}

	level.Debug(c.logger).Log("msg", "archive compression level", "level", c.opts.compressionLevel)

	closer := func() {
		twCloser()
		file.Close()
	}

	defer closer()

	// 3. walk through source and add each file
	err = filepath.Walk(src, writeToArchive(tw, c.opts.skipSymlinks))
	if err != nil {
		return fmt.Errorf("add all files to archive %w", err)
	}

	// 4. Close resources before upload
	closer()

	// 5. upload archive file to server
	level.Info(c.logger).Log("msg", "uploading archived directory", "src", src, "dst", dst)

	return c.pushArchive(dst, archivePath)
}

func (c Cache) pushArchive(dst, archivePath string) error {
	f, err := os.Open(archivePath)
	if err != nil {
		return fmt.Errorf("open archived file to send %w", err)
	}
	defer f.Close()

	if err := c.b.Put(dst, f); err != nil {
		return fmt.Errorf("upload file %w", err)
	}

	return nil
}

// Pull fetches the archived file from the cache and restores to the host machine's file system
func (c Cache) Pull(src, dst string) error {
	level.Info(c.logger).Log("msg", "downloading archived directory", "src", src)
	// 1. download archive
	rc, err := c.b.Get(src)
	if err != nil {
		return fmt.Errorf("get file from storage backend %w", err)
	}
	defer rc.Close()

	// 2. extract archive
	level.Info(c.logger).Log("msg", "extracting archived directory", "src", src, "dst", dst)

	if err := extractFromArchive(archiveReader(rc, c.opts.archiveFmt)); err != nil {
		return fmt.Errorf("extract files from downloaded archive %w", err)
	}

	return nil
}

// Helpers

func archiveWriter(w io.Writer, f string, l int) (*tar.Writer, func(), error) {
	switch f {
	case "gzip":
		gw, err := gzip.NewWriterLevel(w, l)
		if err != nil {
			return nil, nil, fmt.Errorf("create archive writer %w", err)
		}

		tw := tar.NewWriter(gw)

		return tw, func() {
			gw.Close()
			tw.Close()
		}, nil
	default:
		tw := tar.NewWriter(w)
		return tw, func() { tw.Close() }, nil
	}
}

func writeToArchive(tw *tar.Writer, skipSymlinks bool) func(path string, fi os.FileInfo, err error) error {
	return func(path string, fi os.FileInfo, pErr error) error {
		if pErr != nil {
			return pErr
		}

		var h *tar.Header
		// Create header for Regular files and Directories
		var err error
		h, err = tar.FileInfoHeader(fi, "")
		if err != nil {
			return fmt.Errorf("create header for <%s> %w", path, err)
		}

		if isSymlink(fi) {
			if skipSymlinks {
				return nil
			}

			var err error
			if h, err = createSymlinkHeader(fi, path); err != nil {
				return fmt.Errorf("create header for symbolic link %w", err)
			}
		}

		h.Name = path // to give absolute path

		if err := tw.WriteHeader(h); err != nil {
			return fmt.Errorf("write header for <%s> %w", path, err)
		}

		if fi.Mode().IsRegular() { // open and write only if it is a regular file
			if err := writeFileToArchive(tw, path); err != nil {
				return fmt.Errorf("write file to archive %w", err)
			}
		}

		return nil
	}
}

func createSymlinkHeader(fi os.FileInfo, path string) (*tar.Header, error) {
	lnk, err := os.Readlink(path)
	if err != nil {
		return nil, fmt.Errorf("read link <%s> %w", path, err)
	}

	h, err := tar.FileInfoHeader(fi, lnk)
	if err != nil {
		return nil, fmt.Errorf("create symlink header for <%s> %w", path, err)
	}

	return h, nil
}

func writeFileToArchive(tw io.Writer, path string) error {
	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("open file <%s> %w", path, err)
	}
	defer f.Close()

	if _, err := io.Copy(tw, f); err != nil {
		return fmt.Errorf("copy the file <%s> data to the tarball %w", path, err)
	}

	return nil
}

func archiveReader(r io.Reader, archiveFmt string) *tar.Reader {
	tr := tar.NewReader(r)

	switch archiveFmt {
	case "gzip":
		gzr, err := gzip.NewReader(r)
		if err != nil {
			gzr.Close()
			return tr
		}

		return tar.NewReader(gzr)
	default:
		return tr
	}
}

func extractFromArchive(tr *tar.Reader) error {
	for {
		h, err := tr.Next()

		switch {
		case err == io.EOF: // if no more files are found return
			return nil
		case err != nil: // return any other error
			return fmt.Errorf("tar reader failed %w", err)
		case h == nil: // if the header is nil, skip it
			continue
		}

		switch h.Typeflag {
		case tar.TypeDir:
			if err := extractDir(h); err != nil {
				return err
			}

			continue
		case tar.TypeReg, tar.TypeRegA, tar.TypeChar, tar.TypeBlock, tar.TypeFifo:
			if err := extractRegular(h, tr); err != nil {
				return fmt.Errorf("extract regular file %w", err)
			}

			continue
		case tar.TypeSymlink:
			if err := extractSymlink(h); err != nil {
				return fmt.Errorf("extract symbolic link %w", err)
			}

			continue
		case tar.TypeLink:
			if err := extractLink(h); err != nil {
				return fmt.Errorf("extract link %w", err)
			}

			continue
		case tar.TypeXGlobalHeader:
			continue
		default:
			return fmt.Errorf("extract %s, unknown type flag: %c", h.Name, h.Typeflag)
		}
	}
}

func extractDir(h *tar.Header) error {
	if err := os.MkdirAll(h.Name, os.FileMode(h.Mode)); err != nil {
		return fmt.Errorf("create directory <%s> %w", h.Name, err)
	}

	return nil
}

func extractRegular(h *tar.Header, tr io.Reader) error {
	f, err := os.OpenFile(h.Name, os.O_CREATE|os.O_RDWR, os.FileMode(h.Mode))
	if err != nil {
		return fmt.Errorf("open extracted file for writing <%s> %w", h.Name, err)
	}
	defer f.Close()

	if _, err := io.Copy(f, tr); err != nil {
		return fmt.Errorf("copy extracted file for writing <%s> %w", h.Name, err)
	}

	return nil
}

func extractSymlink(h *tar.Header) error {
	if err := unlink(h.Name); err != nil {
		return fmt.Errorf("unlink <%s> %w", h.Name, err)
	}

	if err := os.Symlink(h.Linkname, h.Name); err != nil {
		return fmt.Errorf("create symbolic link <%s> %w", h.Name, err)
	}

	return nil
}

func extractLink(h *tar.Header) error {
	if err := unlink(h.Name); err != nil {
		return fmt.Errorf("unlink <%s> %w", h.Name, err)
	}

	if err := os.Link(h.Linkname, h.Name); err != nil {
		return fmt.Errorf("create hard link <%s> %w", h.Linkname, err)
	}

	return nil
}

func isSymlink(fi os.FileInfo) bool {
	return fi.Mode()&os.ModeSymlink != 0
}

func unlink(path string) error {
	_, err := os.Lstat(path)
	if err == nil {
		return os.Remove(path)
	}

	return nil
}
