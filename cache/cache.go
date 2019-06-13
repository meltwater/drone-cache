// Package cache provides functionality for cache storage
package cache

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
)

// Backend implements operations for caching files
type Backend interface {
	Get(string) (io.ReadCloser, error)
	Put(string, io.ReadSeeker) error
}

// Cache contains configuration for Cache functionality
type Cache struct {
	skipSymlinks bool
	archiveFmt   string
	b            Backend
}

// New creates a new cache with given parameters
func New(b Backend, archiveFmt string, skipSymlinks bool) Cache {
	return Cache{b: b, archiveFmt: archiveFmt, skipSymlinks: skipSymlinks}
}

// Push pushes the archived file to the cache
func (c Cache) Push(src, dst string) error {
	// 1. check if source is reachable
	src, err := filepath.Abs(filepath.Clean(src))
	if err != nil {
		return errors.Wrap(err, "could not read source directory")
	}

	log.Printf("archiving directory <%s>", src)

	// 2. create a temporary file for the archive
	if err := ensureDir("/tmp"); err != nil {
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
	tw, twCloser := archiveWriter(file, c.archiveFmt)
	closer := func() {
		twCloser()
		file.Close()
	}
	defer closer()

	// 3. walk through source and add each file
	err = filepath.Walk(src, writeFileToArchive(tw, c.skipSymlinks))
	if err != nil {
		return errors.Wrap(err, "could not add all files to archive")
	}

	// 4. Close resources before upload
	closer()

	// 5. upload archive file to server
	log.Printf("uploading archived directory <%s> to <%s>", src, dst)
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
	log.Printf("dowloading archived directory <%s>", src)
	// 1. download archive
	rc, err := c.b.Get(src)
	if err != nil {
		return errors.Wrap(err, "could not get file from storage backend")
	}
	defer rc.Close()

	// 2. extract archive
	log.Printf("extracting archived directory <%s> to <%s>", src, dst)
	tr := archiveReader(rc, c.archiveFmt)
	return errors.Wrap(extractFilesFromArchive(tr, dst), "could not extract files from downloaded archive")
}

// Helpers

func archiveWriter(w io.Writer, archiveFmt string) (*tar.Writer, func()) {
	switch archiveFmt {
	case "gzip":
		gw := gzip.NewWriter(w)
		tw := tar.NewWriter(gw)
		return tw, func() {
			gw.Close()
			tw.Close()
		}
	default:
		tw := tar.NewWriter(w)
		return tw, func() { tw.Close() }
	}
}

func writeFileToArchive(tw *tar.Writer, skipSymlinks bool) func(path string, fi os.FileInfo, err error) error {
	return func(path string, fi os.FileInfo, perr error) error {
		if fi.Mode().IsDir() { // skip on directories
			return nil
		}

		var h *tar.Header
		if isSymlink(fi) {
			if skipSymlinks {
				return nil
			}

			linkTarget, err := os.Readlink(path)
			if err != nil {
				return errors.Wrap(err, fmt.Sprintf("could not read link <%s>", path))
			}

			absPath, err := filepath.Abs(filepath.ToSlash(linkTarget))
			if err != nil {
				return errors.Wrap(err, fmt.Sprintf("could not get absolute path for <%s>", linkTarget))
			}

			h, err = tar.FileInfoHeader(fi, absPath)
			if err != nil {
				return errors.Wrap(err, fmt.Sprintf("could not create header for <%s>", path))
			}
		}

		if fi.Mode().IsRegular() {
			h = &tar.Header{
				Name:     path,
				Size:     fi.Size(),
				Mode:     int64(fi.Mode()),
				ModTime:  fi.ModTime(),
				Typeflag: tar.TypeReg,
			}
		}

		if err := tw.WriteHeader(h); err != nil {
			return errors.Wrap(err, fmt.Sprintf("could not write header for <%s>", path))
		}

		if fi.Mode().IsRegular() { // open and write only if it is a regular file
			f, err := os.Open(path)
			if err != nil {
				return errors.Wrap(err, fmt.Sprintf("could not open file <%s>", path))
			}
			defer f.Close()

			if _, err := io.Copy(tw, f); err != nil {
				return errors.Wrap(err, fmt.Sprintf("could not copy the file <%s> data to the tarball", path))
			}
		}

		return nil
	}
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

func extractFilesFromArchive(tr *tar.Reader, dst string) error {
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
			if err := ensureDir(h.Name); err != nil {
				return errors.Wrap(err, fmt.Sprintf("could not create <%s> directory", h.Name))
			}
			continue
		case tar.TypeReg, tar.TypeRegA, tar.TypeChar, tar.TypeBlock, tar.TypeFifo:
			trgt := filepath.Join("/", h.Name)

			dir := filepath.Dir(trgt)
			if err := ensureDir(dir); err != nil {
				return errors.Wrap(err, fmt.Sprintf("could not create <%s> directory", dir))
			}

			f, err := os.OpenFile(trgt, os.O_CREATE|os.O_RDWR, os.FileMode(h.Mode))
			if err != nil {
				return errors.Wrap(err, fmt.Sprintf("could not open extracted file for writing <%s>", trgt))
			}

			if _, err := io.Copy(f, tr); err != nil {
				f.Close()
				return errors.Wrap(err, fmt.Sprintf("could not copy extracted file for writing <%s>", trgt))
			}
			f.Close()
			continue
		case tar.TypeSymlink:
			linkPath := filepath.Join(dst, h.Name)

			dir := filepath.Dir(linkPath)
			if err := ensureDir(dir); err != nil {
				return errors.Wrap(err, fmt.Sprintf("could not create <%s> directory", dir))
			}

			if err := unlink(linkPath); err != nil {
				return errors.Wrap(err, fmt.Sprintf("could not unlink <%s>", linkPath))
			}

			trgt := filepath.Join("/", h.Linkname)
			if err := os.Symlink(trgt, linkPath); err != nil {
				return errors.Wrap(err, fmt.Sprintf("could not create symbolic link <%s>", linkPath))
			}
			continue
		case tar.TypeLink:
			linkPath := filepath.Join(dst, h.Name)

			dir := filepath.Dir(linkPath)
			if err := ensureDir(dir); err != nil {
				return errors.Wrap(err, fmt.Sprintf("could not create <%s> directory", dir))
			}

			if err := unlink(linkPath); err != nil {
				return errors.Wrap(err, fmt.Sprintf("could not unlink <%s>", linkPath))
			}

			trgt := filepath.Join("/", h.Linkname)
			if err := os.Link(trgt, linkPath); err != nil {
				return errors.Wrap(err, fmt.Sprintf("could not create hard link <%s>", linkPath))
			}
			continue
		case tar.TypeXGlobalHeader:
			continue
		default:
			return fmt.Errorf("could not extract %s, unknown type flag: %c", h.Name, h.Typeflag)
		}
	}
}

func ensureDir(dirName string) error {
	if _, err := os.Stat(dirName); err != nil {
		if err := os.MkdirAll(dirName, os.FileMode(0755)); err != nil {
			return errors.Wrap(err, fmt.Sprintf("could not create directory <%s>", dirName))
		}
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
