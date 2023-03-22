package zstd

import (
	"compress/flate"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/go-kit/log"

	"github.com/meltwater/drone-cache/archive/tar"
	"github.com/meltwater/drone-cache/test"
)

var (
	testRoot          = "testdata"
	testRootMounted   = "testdata/mounted"
	testRootExtracted = "testdata/extracted"
)

func TestCreate(t *testing.T) {
	test.Ok(t, os.MkdirAll(testRootMounted, 0755))
	test.Ok(t, os.MkdirAll(testRootExtracted, 0755))
	t.Cleanup(func() { os.RemoveAll(testRoot) })

	for _, tc := range []struct {
		name    string
		tzst    *Archive
		srcs    []string
		written int64
		err     error
	}{
		{
			name:    "empty mount paths",
			tzst:    New(log.NewNopLogger(), testRootMounted, true, flate.DefaultCompression),
			srcs:    []string{},
			written: 0,
			err:     nil,
		},
		{
			name: "non-existing mount paths",
			tzst: New(log.NewNopLogger(), testRootMounted, true, flate.DefaultCompression),
			srcs: []string{
				"iamnotexists",
				"metoo",
			},
			written: 0,
			err:     tar.ErrSourceNotReachable, // os.ErrNotExist || os.ErrPermission
		},
		{
			name:    "existing mount paths",
			tzst:    New(log.NewNopLogger(), testRootMounted, true, flate.DefaultCompression),
			srcs:    exampleFileTree(t, "zstd_create"),
			written: 43, // 3 x tmpfile in dir, 1 tmpfile
			err:     nil,
		},
		{
			name:    "existing mount nested paths",
			tzst:    New(log.NewNopLogger(), testRootMounted, true, flate.DefaultCompression),
			srcs:    exampleNestedFileTree(t, "tar_create"),
			written: 56, // 4 x tmpfile in dir, 1 tmpfile
			err:     nil,
		},
		{
			name:    "existing mount paths with symbolic links",
			tzst:    New(log.NewNopLogger(), testRootMounted, false, flate.DefaultCompression),
			srcs:    exampleFileTreeWithSymlinks(t, "zstd_create_symlink"),
			written: 43,
			err:     nil,
		},
	} {
		tc := tc // NOTE: https://github.com/golang/go/wiki/CommonMistakes#using-goroutines-on-loop-iterator-variables
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Setup
			dstDir, dstDirClean := test.CreateTempDir(t, "zstd_create_archives", testRootMounted)
			t.Cleanup(dstDirClean)

			extDir, extDirClean := test.CreateTempDir(t, "zstd_create_extracted", testRootExtracted)
			t.Cleanup(extDirClean)

			// Run
			archivePath := filepath.Join(dstDir, filepath.Clean(tc.name+".tar.zst"))
			written, err := create(tc.tzst, tc.srcs, archivePath)
			if err != nil {
				test.Expected(t, err, tc.err)
				return
			}

			test.Exists(t, archivePath)
			test.Assert(t, written == tc.written, "case %q: written bytes got %d want %v", tc.name, written, tc.written)

			_, err = extract(tc.tzst, archivePath, extDir)
			test.Ok(t, err)
			test.EqualDirs(t, extDir, testRootMounted, tc.srcs)
		})
	}
}

func TestExtract(t *testing.T) {
	test.Ok(t, os.MkdirAll(testRootMounted, 0755))
	test.Ok(t, os.MkdirAll(testRootExtracted, 0755))
	t.Cleanup(func() { os.RemoveAll(testRoot) })

	// Setup
	tzst := New(log.NewNopLogger(), testRootMounted, false, flate.DefaultCompression)

	arcDir, arcDirClean := test.CreateTempDir(t, "zstd_extract_archive")
	t.Cleanup(arcDirClean)

	files := exampleFileTree(t, "zstd_extract")
	archivePath := filepath.Join(arcDir, "test.tar.zst")
	_, err := create(tzst, files, archivePath)
	test.Ok(t, err)

	nestedFiles := exampleNestedFileTree(t, "zstd_extract_nested")
	nestedArchivePath := filepath.Join(arcDir, "nested_test.tar.zst")
	_, err = create(tzst, nestedFiles, nestedArchivePath)
	test.Ok(t, err)

	filesWithSymlink := exampleFileTreeWithSymlinks(t, "zstd_extract_symlink")
	archiveWithSymlinkPath := filepath.Join(arcDir, "test_with_symlink.tar.zst")
	_, err = create(tzst, filesWithSymlink, archiveWithSymlinkPath)
	test.Ok(t, err)

	emptyArchivePath := filepath.Join(arcDir, "empty_test.tar.zst")
	_, err = create(tzst, []string{}, emptyArchivePath)
	test.Ok(t, err)

	badArchivePath := filepath.Join(arcDir, "bad_test.tar.zst")
	test.Ok(t, ioutil.WriteFile(badArchivePath, []byte("hello\ndrone\n"), 0644))

	for _, tc := range []struct {
		name        string
		tzst        *Archive
		archivePath string
		srcs        []string
		written     int64
		err         error
	}{
		{
			name:        "non-existing archive",
			tzst:        New(log.NewNopLogger(), testRootMounted, true, flate.DefaultCompression),
			archivePath: "iamnotexists",
			srcs:        []string{},
			written:     0,
			err:         os.ErrNotExist,
		},
		{
			name:        "non-existing root destination",
			tzst:        New(log.NewNopLogger(), testRootMounted, true, flate.DefaultCompression),
			archivePath: emptyArchivePath,
			srcs:        []string{},
			written:     0,
			err:         nil,
		},
		{
			name:        "empty archive",
			tzst:        New(log.NewNopLogger(), testRootMounted, true, flate.DefaultCompression),
			archivePath: emptyArchivePath,
			srcs:        []string{},
			written:     0,
			err:         nil,
		},
		{
			name:        "bad archives",
			tzst:        New(log.NewNopLogger(), testRootMounted, true, flate.DefaultCompression),
			archivePath: badArchivePath,
			srcs:        []string{},
			written:     0,
			err:         tar.ErrArchiveNotReadable,
		},
		{
			name:        "existing archive",
			tzst:        New(log.NewNopLogger(), testRootMounted, true, flate.DefaultCompression),
			archivePath: archivePath,
			srcs:        files,
			written:     43,
			err:         nil,
		},
		{
			name:        "existing archive with nested files",
			tzst:        New(log.NewNopLogger(), testRootMounted, true, flate.DefaultCompression),
			archivePath: nestedArchivePath,
			srcs:        nestedFiles,
			written:     56,
			err:         nil,
		},
		{
			name:        "existing archive with symbolic links",
			tzst:        New(log.NewNopLogger(), testRootMounted, false, flate.DefaultCompression),
			archivePath: archiveWithSymlinkPath,
			srcs:        filesWithSymlink,
			written:     43,
			err:         nil,
		},
	} {
		tc := tc // NOTE: https://github.com/golang/go/wiki/CommonMistakes#using-goroutines-on-loop-iterator-variables
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			dstDir, dstDirClean := test.CreateTempDir(t, "zstd_extract_"+tc.name, testRootExtracted)
			t.Cleanup(dstDirClean)

			written, err := extract(tc.tzst, tc.archivePath, dstDir)
			if err != nil {
				test.Expected(t, err, tc.err)
				return
			}

			test.Assert(t, written == tc.written, "case %q: written bytes got %d want %v", tc.name, written, tc.written)
			test.EqualDirs(t, dstDir, testRootMounted, tc.srcs)
		})
	}
}

// Helpers

func create(a *Archive, srcs []string, dst string) (int64, error) {
	pr, pw := io.Pipe()
	defer pr.Close()

	var written int64
	go func(w *int64) {
		defer pw.Close()

		written, err := a.Create(srcs, pw, false)
		if err != nil {
			pw.CloseWithError(err)
		}

		*w = written
	}(&written)

	content, err := ioutil.ReadAll(pr)
	if err != nil {
		pr.CloseWithError(err)
		return 0, err
	}

	if err := ioutil.WriteFile(dst, content, 0644); err != nil {
		return 0, err
	}

	return written, nil
}

func extract(a *Archive, src string, dst string) (int64, error) {
	pr, pw := io.Pipe()
	defer pr.Close()

	f, err := os.Open(src)
	if err != nil {
		return 0, err
	}

	go func() {
		defer pw.Close()

		_, err = io.Copy(pw, f)
		if err != nil {
			pw.CloseWithError(err)
		}
	}()

	return a.Extract(dst, pr)
}

// Fixtures

func exampleFileTree(t *testing.T, name string) []string {
	file, fileClean := test.CreateTempFile(t, name, []byte("hello\ndrone!\n"), testRootMounted) // 13 bytes
	t.Cleanup(fileClean)

	dir, dirClean := test.CreateTempFilesInDir(t, name, []byte("hello\ngo!\n"), testRootMounted) // 10 bytes
	t.Cleanup(dirClean)

	return []string{file, dir}
}

func exampleNestedFileTree(t *testing.T, name string) []string {
	dir, cleanup := test.CreateTempDir(t, name, testRootMounted)
	t.Cleanup(cleanup)

	nestedFile, nestedFileClean := test.CreateTempFile(t, name, []byte("hello\ndrone!\n"), dir) // 13 bytes
	t.Cleanup(nestedFileClean)

	nestedDir, nestedDirClean := test.CreateTempFilesInDir(t, name, []byte("hello\ngo!\n"), dir) // 10 bytes
	t.Cleanup(nestedDirClean)

	nestedDir1, nestedDirClean1 := test.CreateTempDir(t, name, dir)
	t.Cleanup(nestedDirClean1)

	nestedDir2, nestedDirClean2 := test.CreateTempDir(t, name, nestedDir1)
	t.Cleanup(nestedDirClean2)

	nestedFile1, nestedFileClean1 := test.CreateTempFile(t, name, []byte("hello\ndrone!\n"), nestedDir2) // 13 bytes
	t.Cleanup(nestedFileClean1)

	return []string{nestedDir, nestedFile, nestedFile1}
}

func exampleFileTreeWithSymlinks(t *testing.T, name string) []string {
	file, fileClean := test.CreateTempFile(t, name, []byte("hello\ndrone!\n"), testRootMounted) // 13 bytes
	t.Cleanup(fileClean)

	symlink := filepath.Join(filepath.Dir(file), name+"_symlink.testfile")
	test.Ok(t, os.Symlink(file, symlink))
	t.Cleanup(func() { os.Remove(symlink) })

	dir, dirClean := test.CreateTempFilesInDir(t, name, []byte("hello\ngo!\n"), testRootMounted) // 10 bytes
	t.Cleanup(dirClean)

	return []string{file, dir, symlink}
}
