// Package cache provides functionality for cache storage
package cache

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
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
	tw, closer := archiveWriter(file, c.archiveFmt)

	// 3. walk through source and add each file
	err = filepath.Walk(src, writeFileToArchive(tw, src))
	if err != nil {
		closer()
		file.Close()
		return errors.Wrap(err, "could not add all files to archive")
	}

	// 4. Close resources before upload
	closer()
	file.Close()

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
	rc, err := c.b.Get(src)
	if err != nil {
		return errors.Wrap(err, "could not get file from storage backend")
	}
	defer rc.Close()

	// create temp file for archive
	temp, err := ioutil.TempFile("", "")
	if err != nil {
		return errors.Wrap(err, "could not create tmp file to archive")
	}
	defer func() {
		temp.Close()
		os.Remove(temp.Name())
	}()

	// download archive to temp file
	if _, err := io.Copy(temp, rc); err != nil {
		errors.Wrap(err, "could not copy downloaded file to tmp")
	}

	// run extraction command
	log.Printf("extracting archived directory <%s> to <%s>", src, dst)
	cmd := exec.Command("tar", "-xf", temp.Name(), "-C", "/")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	defer log.Printf("command stout: <%s>, stderr: <%s>", string(stdout.Bytes()), string(stderr.Bytes()))

	return errors.Wrap(cmd.Run(), "could not open extract downloaded file")
}

// Helpers

func archiveWriter(w io.Writer, archiveFmt string) (*tar.Writer, func()) {
	var tw *tar.Writer
	var closer func()
	switch archiveFmt {
	case "tar":
		tw = tar.NewWriter(w)
		closer = func() { tw.Close() }
	case "gzip":
		gw := gzip.NewWriter(w)
		tw = tar.NewWriter(gw)
		closer = func() {
			gw.Close()
			tw.Close()
		}
	default:
		tw = tar.NewWriter(w)
		closer = func() { tw.Close() }
	}
	return tw, closer
}

func writeFileToArchive(tw *tar.Writer, rootPath string) func(path string, fi os.FileInfo, err error) error {
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
