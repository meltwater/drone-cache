//nolint:gomnd
package test

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"testing"

	"github.com/google/go-cmp/cmp"
)

// Assert is modified version of https://github.com/benbjohnson/testing.

// Assert fails the test if the condition is false.
func Assert(tb testing.TB, condition bool, msg string, v ...interface{}) {
	tb.Helper()

	if !condition {
		_, file, line, _ := runtime.Caller(1)
		tb.Fatalf("%s:%d: "+msg+"\n", append([]interface{}{filepath.Base(file), line}, v...)...)
	}
}

// Ok fails the test if an err is not nil.
func Ok(tb testing.TB, err error) {
	tb.Helper()

	if err != nil {
		_, file, line, _ := runtime.Caller(1)
		tb.Fatalf("%s:%d: unexpected error: %s\n", filepath.Base(file), line, err.Error())
	}
}

// NotOk fails the test if an err is nil.
func NotOk(tb testing.TB, err error) {
	tb.Helper()

	if err == nil {
		_, file, line, _ := runtime.Caller(1)
		tb.Fatalf("%s:%d: expected error, got nothing\n", filepath.Base(file), line)
	}
}

// Expected fails if the errors does not match.
func Expected(tb testing.TB, got, want error) {
	tb.Helper()

	NotOk(tb, got)

	if errors.Is(got, want) {
		return
	}

	_, file, line, _ := runtime.Caller(1)
	tb.Fatalf("%s:%d: got unexpected error: %v\n", filepath.Base(file), line, got.Error())
}

// Exists fails if the file or directory in the given path does not exist.
func Exists(tb testing.TB, path string) {
	tb.Helper()

	_, err := os.Lstat(path)
	if err != nil {
		_, file, line, _ := runtime.Caller(1)

		if os.IsNotExist(err) {
			tb.Fatalf("%s:%d: should exists: %s\n", filepath.Base(file), line, err.Error())
		}
	}
}

// Equals fails the test if want is not equal to got.
func Equals(tb testing.TB, want, got interface{}, v ...interface{}) {
	tb.Helper()

	if diff := cmp.Diff(want, got); diff != "" {
		_, file, line, _ := runtime.Caller(1)

		var msg string

		if len(v) > 0 {
			msg = fmt.Sprintf(v[0].(string), v[1:]...)
		}

		tb.Fatalf("%s:%d:"+msg+"\n\n\t (-want +got):\n%s", filepath.Base(file), line, diff)
	}
}

//nolint:funlen // EqualDirs fails if the contents of given directories are not the same.
func EqualDirs(tb testing.TB, dst string, src string, srcs []string) {
	tb.Helper()

	srcList := []string{}

	for _, s := range srcs {
		if isDir(s) {
			paths, err := expand(s)
			if err != nil {
				tb.Fatalf("expand %s: %v\n", s, err)
			}

			srcList = append(srcList, paths...)

			continue
		}

		srcList = append(srcList, s)
	}

	dstList, err := expand(dst)
	if err != nil {
		tb.Fatalf("expand %s: %v\n", dst, err)
	}

	sort.Strings(srcList)
	sort.Strings(dstList)

	relSrcList, err := relative(src, srcList)
	if err != nil {
		tb.Fatalf("relative %s: %v\n", src, err)
	}

	relDstList, err := relative(dst, dstList)
	if err != nil {
		tb.Fatalf("relative %s: %v\n", dst, err)
	}

	Equals(tb, relSrcList, relDstList)

	_, file, line, _ := runtime.Caller(1)

	for i := 0; i < len(srcList); i++ {
		src := srcList[i]
		dst := dstList[i]

		if isSymlink(src) && isSymlink(dst) {
			src, err = os.Readlink(src)
			if err != nil {
				tb.Fatalf("%s:%d: unexpected error, src path, link <%s>: %s\n",
					filepath.Base(file), line, src, err.Error())
			}

			dst, err = os.Readlink(dst)
			if err != nil {
				tb.Fatalf("%s:%d: unexpected error, dst path, link <%s>: %s\n",
					filepath.Base(file), line, dst, err.Error())
			}
		}

		wContent, err := ioutil.ReadFile(src)
		if err != nil {
			tb.Fatalf("%s:%d: unexpected error, src path <%s>: %s\n", filepath.Base(file), line, srcList[i], err.Error())
		}

		gContent, err := ioutil.ReadFile(dst)
		if err != nil {
			tb.Fatalf("%s:%d: unexpected error, dst path <%s>: %s\n",
				filepath.Base(file), line, dstList[i], err.Error())
		}

		Equals(tb, wContent, gContent)
	}
}

// Helpers

func isDir(path string) bool {
	fi, err := os.Stat(path)
	if err != nil {
		return false
	}

	return fi.IsDir()
}

func isSymlink(path string) bool {
	fi, err := os.Lstat(path)
	return err == nil && fi.Mode()&os.ModeSymlink != 0
}

func expand(src string) ([]string, error) {
	paths := []string{}

	if err := filepath.Walk(src, func(path string, fi os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("walk %q: %v", path, err)
		}

		if fi.IsDir() {
			return nil
		}

		paths = append(paths, path)

		return nil
	}); err != nil {
		return nil, fmt.Errorf("walking the path %q: %v", src, err)
	}

	return paths, nil
}

func relative(top string, paths []string) ([]string, error) {
	result := make([]string, len(paths))

	for _, p := range paths {
		name := filepath.Base(p)

		rel, err := filepath.Rel(top, filepath.Dir(p))
		if err != nil {
			return []string{}, fmt.Errorf("relative path %q: %q %v", p, rel, err)
		}

		name = filepath.Join(filepath.ToSlash(rel), name)
		result = append(result, name)
	}

	return result, nil
}
