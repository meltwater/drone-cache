// Package cache provides functionality for cache storage
package cache

import (
	"bytes"
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

// Upload pushes the archived file to the cache
func Upload(b Backend, src, dst string) error {
	var err error
	src, err = filepath.Abs(filepath.Clean(src))
	if err != nil {
		return errors.Wrap(err, "could not read source directory")
	}

	// create a temporary file for the archive
	dir, err := ioutil.TempDir("", "")
	if err != nil {
		return errors.Wrap(err, "could not create tmp folder to archive")
	}
	tar := filepath.Join(dir, "archive.tar")

	// run archive command
	log.Printf("archiving directory <%s>", src)
	cmd := exec.Command("tar", "-cf", tar, src)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return errors.Wrap(err, "external command (tart) run failed")
	}
	log.Printf("command stout: <%s>, stderr: <%s>", string(stdout.Bytes()), string(stderr.Bytes()))

	// upload file to server
	f, err := os.Open(tar)
	if err != nil {
		return errors.Wrap(err, "could not open archived file to send")
	}
	defer f.Close()

	log.Printf("uploading archived directory <%s> to <%s>", src, dst)
	return errors.Wrap(b.Put(dst, f), "could not upload file")
}

// Download fetches the archived file from the cache
// and restores to the host machine's file system
func Download(b Backend, src, dst string) error {
	log.Printf("dowloading archived directory <%s>", src)
	rc, err := b.Get(src)
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
