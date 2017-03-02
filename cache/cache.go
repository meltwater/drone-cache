package cache

import (
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
)

// Cache implements operations for caching files.
type Cache interface {
	Get(string) (io.ReadCloser, error)
	Put(string, io.ReadSeeker) error
}

// RebuildCmd is a helper function that pushes the archived file to the cache.
func RebuildCmd(c Cache, src, dst string) (err error) {

	src = filepath.Clean(src)
	src, err = filepath.Abs(src)
	if err != nil {
		return err
	}

	// create a temporary file for the archive
	dir, err := ioutil.TempDir("", "")
	if err != nil {
		return err
	}
	tar := filepath.Join(dir, "archive.tar")

	// run archive command
	cmd := exec.Command("tar", "-cf", tar, src)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}

	// upload file to server
	f, err := os.Open(tar)
	if err != nil {
		return err
	}
	defer f.Close()
	return c.Put(dst, f)
}

// RestoreCmd is a helper function that fetches the archived file from the cache
// and restores to the host machine's file system.
func RestoreCmd(c Cache, src, dst string) error {
	rc, err := c.Get(src)
	if err != nil {
		return err
	}
	defer rc.Close()

	// create temp file for archive
	temp, err := ioutil.TempFile("", "")
	if err != nil {
		return err
	}
	defer func() {
		temp.Close()
		os.Remove(temp.Name())
	}()

	// download archive to temp file
	if _, err := io.Copy(temp, rc); err != nil {
		return err
	}

	// cleanup after ourself
	temp.Close()

	// run extraction command
	cmd := exec.Command("tar", "-xf", temp.Name(), "-C", "/")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
