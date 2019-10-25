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
	"github.com/pkg/errors"
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
		return errors.Wrap(err, "could not read source directory")
	}

	level.Info(c.logger).Log("msg", "archiving directory", "src", src)

	// 2. create a temporary file for the archive
	if err := os.MkdirAll("/tmp", os.FileMode(0755)); err != nil {
		return errors.Wrap(err, "could not create tmp directory")
	}

	dir, err := ioutil.TempDir("", "")
	if err != nil {
		return errors.Wrap(err, "could not create tmp folder for archive")
	}

	archivePath := filepath.Join(dir, "archive.tar")

	file, err := os.Create(archivePath)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("could not create tarball file <%s>", archivePath))
	}

	tw, twCloser, err := archiveWriter(file, c.opts.archiveFmt, c.opts.compressionLevel)
	if err != nil {
		return errors.Wrap(err, "could not initialize archive writer")
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
		return errors.Wrap(err, "could not add all files to archive")
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
		return errors.Wrap(err, "could not open archived file to send")
	}
	defer f.Close()

	return errors.Wrap(c.b.Put(dst, f), "could not upload file")
}

// Pull fetches the archived file from the cache and restores to the host machine's file system
func (c Cache) Pull(src, dst string) error {
	level.Info(c.logger).Log("msg", "downloading archived directory", "src", src)
	// 1. download archive
	rc, err := c.b.Get(src)
	if err != nil {
		return errors.Wrap(err, "could not get file from storage backend")
	}
	defer rc.Close()

	// 2. extract archive
	level.Info(c.logger).Log("msg", "extracting archived directory", "src", src, "dst", dst)

	return errors.Wrap(
		extractFromArchive(archiveReader(rc, c.opts.archiveFmt)),
		"could not extract files from downloaded archive",
	)
}

// Helpers

func archiveWriter(w io.Writer, fmt string, l int) (*tar.Writer, func(), error) {
	switch fmt {
	case "gzip":
		gw, err := gzip.NewWriterLevel(w, l)
		if err != nil {
			return nil, nil, errors.Wrap(err, "could not create archive writer")
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
			return errors.Wrap(err, fmt.Sprintf("could not create header for <%s>", path))
		}

		if isSymlink(fi) {
			if skipSymlinks {
				return nil
			}

			var err error
			if h, err = createSymlinkHeader(fi, path); err != nil {
				return errors.Wrap(err, "could not create header for symbolic link")
			}
		}

		h.Name = path // to give absolute path

		if err := tw.WriteHeader(h); err != nil {
			return errors.Wrap(err, fmt.Sprintf("could not write header for <%s>", path))
		}

		if fi.Mode().IsRegular() { // open and write only if it is a regular file
			if err := writeFileToArchive(tw, path); err != nil {
				return errors.Wrap(err, "could not write file to archive")
			}
		}

		return nil
	}
}

func createSymlinkHeader(fi os.FileInfo, path string) (*tar.Header, error) {
	lnk, err := os.Readlink(path)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("could not read link <%s>", path))
	}

	h, err := tar.FileInfoHeader(fi, lnk)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("could not create symlink header for <%s>", path))
	}

	return h, nil
}

func writeFileToArchive(tw io.Writer, path string) error {
	f, err := os.Open(path)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("could not open file <%s>", path))
	}
	defer f.Close()

	if _, err := io.Copy(tw, f); err != nil {
		return errors.Wrap(err, fmt.Sprintf("could not copy the file <%s> data to the tarball", path))
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
			return errors.Wrap(err, "tar reader failed")
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
				return errors.Wrap(err, "could not extract regular file")
			}

			continue
		case tar.TypeSymlink:
			if err := extractSymlink(h); err != nil {
				return errors.Wrap(err, "could not extract symbolic link")
			}

			continue
		case tar.TypeLink:
			if err := extractLink(h); err != nil {
				return errors.Wrap(err, "could not extract link")
			}

			continue
		case tar.TypeXGlobalHeader:
			continue
		default:
			return fmt.Errorf("could not extract %s, unknown type flag: %c", h.Name, h.Typeflag)
		}
	}
}

func extractDir(h *tar.Header) error {
	if err := os.MkdirAll(h.Name, os.FileMode(h.Mode)); err != nil {
		return errors.Wrap(err, fmt.Sprintf("could not create directory <%s>", h.Name))
	}

	return nil
}

func extractRegular(h *tar.Header, tr io.Reader) error {
	f, err := os.OpenFile(h.Name, os.O_CREATE|os.O_RDWR, os.FileMode(h.Mode))
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("could not open extracted file for writing <%s>", h.Name))
	}
	defer f.Close()

	if _, err := io.Copy(f, tr); err != nil {
		return errors.Wrap(err, fmt.Sprintf("could not copy extracted file for writing <%s>", h.Name))
	}

	return nil
}

func extractSymlink(h *tar.Header) error {
	if err := unlink(h.Name); err != nil {
		return errors.Wrap(err, fmt.Sprintf("could not unlink <%s>", h.Name))
	}

	if err := os.Symlink(h.Linkname, h.Name); err != nil {
		return errors.Wrap(err, fmt.Sprintf("could not create symbolic link <%s>", h.Name))
	}

	return nil
}

func extractLink(h *tar.Header) error {
	if err := unlink(h.Name); err != nil {
		return errors.Wrap(err, fmt.Sprintf("could not unlink <%s>", h.Name))
	}

	if err := os.Link(h.Linkname, h.Name); err != nil {
		return errors.Wrap(err, fmt.Sprintf("could not create hard link <%s>", h.Linkname))
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
