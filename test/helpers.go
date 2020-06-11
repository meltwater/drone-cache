package test

import (
	"io/ioutil"
	"os"
	"testing"
)

// CreateTempFile is a test helper to create a temporary file with given name and content, in given directory.
func CreateTempFile(t testing.TB, name string, content []byte, in ...string) (string, func()) {
	t.Helper()

	parent := ""
	if len(in) > 0 {
		parent = in[0]
	}

	tmpfile, err := ioutil.TempFile(parent, name+"_*.testfile")
	if err != nil {
		t.Fatalf("unexpectedly failed creating the temp file: %v", err)
	}

	if _, err := tmpfile.Write(content); err != nil {
		t.Fatalf("unexpectedly failed writing to the temp file: %v", err)
	}

	if err := tmpfile.Close(); err != nil {
		t.Fatalf("unexpectedly failed closing the temp file: %v", err)
	}

	return tmpfile.Name(), func() { os.Remove(tmpfile.Name()) }
}

// CreateTempFilesInDir is a test helper to create temporary files using name as prefix and content, in given directory.
func CreateTempFilesInDir(t testing.TB, name string, content []byte, in ...string) (string, func()) {
	t.Helper()

	parent := ""
	if len(in) > 0 {
		parent = in[0]
	}

	tmpDir, err := ioutil.TempDir(parent, name+"-testdir-*")
	if err != nil {
		t.Fatalf("unexpectedly failed creating the temp dir: %v", err)
	}

	for i := 0; i < 3; i++ {
		tmpfile, err := ioutil.TempFile(tmpDir, name+"_*.testfile")
		if err != nil {
			t.Fatalf("unexpectedly failed creating the temp file: %v", err)
		}

		if _, err := tmpfile.Write(content); err != nil {
			t.Fatalf("unexpectedly failed writing to the temp file: %v", err)
		}

		if err := tmpfile.Close(); err != nil {
			t.Fatalf("unexpectedly failed closing the temp file: %v", err)
		}
	}

	return tmpDir, func() { os.RemoveAll(tmpDir) }
}

// CreateTempDir is a test helper to create a temporary directory, in given directory.
func CreateTempDir(t testing.TB, name string, in ...string) (string, func()) {
	t.Helper()

	parent := ""
	if len(in) > 0 {
		parent = in[0]
	}

	tmpDir, err := ioutil.TempDir(parent, name+"-testdir-*")
	if err != nil {
		t.Fatalf("unexpectedly failed creating the temp dir: %v", err)
	}

	return tmpDir, func() { os.RemoveAll(tmpDir) }
}
