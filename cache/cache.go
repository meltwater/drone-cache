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
	b          Backend
	archiveFmt string
}

// New creates a new cache with given parameters
func New(b Backend, archiveFmt string) Cache {
	return Cache{b: b, archiveFmt: archiveFmt}
}

// Upload pushes the archived file to the cache
func (c Cache) Upload(src, dst string) error {
	// 1. check if source is reachable
	src, err := filepath.Abs(filepath.Clean(src))
	if err != nil {
		return errors.Wrap(err, "could not read source directory")
	}

	log.Printf("archiving directory <%s>", src)

	// 2. create a temporary file for the archive
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
	err = filepath.Walk(src, writeFileToArchive(tw, src))
	if err != nil {
		return errors.Wrap(err, "could not add all files to archive")
	}

	// 4. Close resources before upload
	closer()

	// 5. upload archive file to server
	log.Printf("uploading archived directory <%s> to <%s>", src, dst)
	return c.uploadArchive(dst, archivePath)
}

func (c Cache) uploadArchive(dst, archivePath string) error {
	f, err := os.Open(archivePath)
	if err != nil {
		return errors.Wrap(err, "could not open archived file to send")
	}
	defer f.Close()

	return errors.Wrap(c.b.Put(dst, f), "could not upload file")
}

// Download fetches the archived file from the cache and restores to the host machine's file system
func (c Cache) Download(src, dst string) error {
	log.Printf("dowloading archived directory <%s>", src)

	// 1. download archive
	rc, err := c.b.Get(src)
	if err != nil {
		return errors.Wrap(err, "could not get file from storage backend")
	}
	defer rc.Close()

	// 2. create temp file for archive
	temp, err := ioutil.TempFile("", "")
	if err != nil {
		return errors.Wrap(err, "could not create tmp file to archive")
	}
	defer func() {
		temp.Close()
		os.Remove(temp.Name())
	}()

	// 3. write downloaded archive to temp file
	if _, err := io.Copy(temp, rc); err != nil {
		errors.Wrap(err, "could not copy downloaded file to tmp")
	}
	rc.Close()

	// 4. extract archive
	log.Printf("extracting archived directory <%s> to <%s>", src, dst)
	tr := archiveReader(temp, c.archiveFmt)
	return errors.Wrap(extractFilesFromArchive(tr, dst), "could not extract files from downloaded archive")

	// // run extraction command
	// cmd := exec.Command("tar", "-xf", temp.Name(), "-C", "/")
	// var stdout, stderr bytes.Buffer
	// cmd.Stdout = &stdout
	// cmd.Stderr = &stderr
	// defer log.Printf("command stout: <%s>, stderr: <%s>", string(stdout.Bytes()), string(stderr.Bytes()))

	// return errors.Wrap(cmd.Run(), "could not open extract downloaded file")
}

// Helpers

func archiveWriter(w io.Writer, archiveFmt string) (*tar.Writer, func()) {
	tw := tar.NewWriter(w)
	closer := func() { tw.Close() }
	if archiveFmt == "gzip" {
		gw := gzip.NewWriter(w)
		tw = tar.NewWriter(gw)
		closer = func() {
			gw.Close()
			tw.Close()
		}
	}
	return tw, closer
}

func writeFileToArchive(tw *tar.Writer, src string) func(path string, fi os.FileInfo, err error) error {
	return func(path string, fi os.FileInfo, err error) error {
		if !fi.Mode().IsRegular() { // skip on symbolic links or directories
			return nil
		}

		f, err := os.Open(path)
		if err != nil {
			return errors.Wrap(err, fmt.Sprintf("could not open file <%s>", path))
		}
		defer f.Close()

		h := &tar.Header{
			Name:    path,
			Size:    fi.Size(),
			Mode:    int64(fi.Mode()),
			ModTime: fi.ModTime(),
		}

		err = tw.WriteHeader(h)
		if err != nil {
			return errors.Wrap(err, fmt.Sprintf("could not write header for file <%s>", path))
		}

		if _, err := io.Copy(tw, f); err != nil {
			return errors.Wrap(err, fmt.Sprintf("could not copy the file <%s> data to the tarball", path))
		}

		return nil
	}
}

func archiveReader(r io.Reader, archiveFmt string) *tar.Reader {
	tr := tar.NewReader(r)
	if archiveFmt == "gzip" {
		gzr, err := gzip.NewReader(r)
		if err != nil {
			gzr.Close()
			return tr
		}
		return tar.NewReader(gzr)
	}
	return tr
}

func extractFilesFromArchive(tr *tar.Reader, dst string) error {
	for {
		h, err := tr.Next()
		switch {
		case err == io.EOF: // if no more files are found return
			return nil
		case err != nil: // return any other error
			return err
		case h == nil: // if the header is nil,skip it
			continue
		}

		target := filepath.Join(dst, h.Name) // the target location where the dir/file should be created

		// the following switch could also be done using fi.Mode(), not sure if there
		// a benefit of using one vs. the other.
		// fi := header.FileInfo()

		switch h.Typeflag {
		case tar.TypeDir: // if its a dir and it doesn't exist create it
			if _, err := os.Stat(target); err != nil {
				if err := os.MkdirAll(target, 0755); err != nil {
					return err
				}
			}

		case tar.TypeReg: // if it's a file create it
			f, err := os.OpenFile(target, os.O_CREATE|os.O_RDWR, os.FileMode(h.Mode))
			if err != nil {
				return err
			}
			if _, err := io.Copy(f, tr); err != nil {
				f.Close()
				return err
			}

			f.Close()
		}
	}
}
