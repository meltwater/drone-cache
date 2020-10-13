package tar

import (
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/meltwater/drone-cache/test"

	"github.com/go-kit/kit/log"
)

var (
	testRoot          = "testdata"
	testRootMounted   = "testdata/mounted"
	testRootExtracted = "testdata/extracted"
	testAbsPattern    = "testdata_absolute"
)

func TestCreate(t *testing.T) {
	test.Ok(t, os.MkdirAll(testRootMounted, 0755))
	test.Ok(t, os.MkdirAll(testRootExtracted, 0755))

	testAbs, err := ioutil.TempDir("", testAbsPattern)
	test.Ok(t, err)
	test.Equals(t, filepath.IsAbs(testAbs), true)

	t.Cleanup(func() {
		os.RemoveAll(testRoot)
		os.RemoveAll(testAbs)
	})

	for _, tc := range []struct {
		name    string
		ta      *Archive
		srcs    []string
		written int64
		err     error
	}{
		{
			name:    "empty mount paths",
			ta:      New(log.NewNopLogger(), testRootMounted, true),
			srcs:    []string{},
			written: 0,
			err:     nil,
		},
		{
			name: "non-existing mount paths",
			ta:   New(log.NewNopLogger(), testRootMounted, true),
			srcs: []string{
				"idonotexist",
				"metoo",
			},
			written: 0,
			err:     ErrSourceNotReachable, // os.ErrNotExist || os.ErrPermission
		},
		{
			name:    "existing mount paths",
			ta:      New(log.NewNopLogger(), testRootMounted, true),
			srcs:    exampleFileTree(t, "tar_create", testRootMounted),
			written: 43, // 3 x tmpfile in dir, 1 tmpfile
			err:     nil,
		},
		{
			name:    "existing mount nested paths",
			ta:      New(log.NewNopLogger(), testRootMounted, true),
			srcs:    exampleNestedFileTree(t, "tar_create"),
			written: 56, // 4 x tmpfile in dir, 1 tmpfile
			err:     nil,
		},
		{
			name:    "existing mount paths with symbolic links",
			ta:      New(log.NewNopLogger(), testRootMounted, false),
			srcs:    exampleFileTreeWithSymlinks(t, "tar_create_symlink"),
			written: 43,
			err:     nil,
		},
		{
			name:    "absolute mount paths",
			ta:      New(log.NewNopLogger(), testRootMounted, true),
			srcs:    exampleFileTree(t, "tar_create", testAbs),
			written: 43,
			err:     nil,
		},
	} {
		tc := tc // NOTICE: https://github.com/golang/go/wiki/CommonMistakes#using-goroutines-on-loop-iterator-variables
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Setup
			var absSrcs []string
			var relativeSrcs []string

			for _, src := range tc.srcs {
				if strings.HasPrefix(src, "/") {
					absSrcs = append(absSrcs, src)
				} else {
					relativeSrcs = append(relativeSrcs, src)
				}
			}

			dstDir, dstDirClean := test.CreateTempDir(t, "tar_create_archives", testRootMounted)
			t.Cleanup(dstDirClean)

			extDir, extDirClean := test.CreateTempDir(t, "tar_create_extracted", testRootExtracted)
			t.Cleanup(extDirClean)

			// Run
			archivePath := filepath.Join(dstDir, filepath.Clean(tc.name+".tar"))
			written, err := create(tc.ta, tc.srcs, archivePath)
			if err != nil {
				test.Expected(t, err, tc.err)
				return
			}

			// Test
			for _, src := range absSrcs {
				test.Ok(t, os.RemoveAll(src))
			}

			test.Exists(t, archivePath)
			test.Assert(t, written == tc.written, "case %q: written bytes got %d want %v", tc.name, written, tc.written)

			_, err = extract(tc.ta, archivePath, extDir)
			test.Ok(t, err)
			test.EqualDirs(t, extDir, testRootMounted, relativeSrcs)

			for _, src := range absSrcs {
				test.Exists(t, src)
			}
		})
	}
}

func TestExtract(t *testing.T) {
	test.Ok(t, os.MkdirAll(testRootMounted, 0755))
	test.Ok(t, os.MkdirAll(testRootExtracted, 0755))

	testAbs, err := ioutil.TempDir("", testAbsPattern)
	test.Ok(t, err)
	test.Equals(t, filepath.IsAbs(testAbs), true)

	t.Cleanup(func() {
		os.RemoveAll(testRoot)
		os.RemoveAll(testAbs)
	})

	// Setup
	ta := New(log.NewNopLogger(), testRootMounted, false)

	arcDir, arcDirClean := test.CreateTempDir(t, "tar_extract_archives", testRootMounted)
	t.Cleanup(arcDirClean)

	files := exampleFileTree(t, "tar_extract", testRootMounted)
	archivePath := filepath.Join(arcDir, "test.tar")
	_, err = create(ta, files, archivePath)
	test.Ok(t, err)

	nestedFiles := exampleNestedFileTree(t, "tar_extract_nested")
	nestedArchivePath := filepath.Join(arcDir, "nested_test.tar")
	_, err = create(ta, nestedFiles, nestedArchivePath)
	test.Ok(t, err)

	filesWithSymlink := exampleFileTreeWithSymlinks(t, "tar_extract_symlink")
	archiveWithSymlinkPath := filepath.Join(arcDir, "test_with_symlink.tar")
	_, err = create(ta, filesWithSymlink, archiveWithSymlinkPath)
	test.Ok(t, err)

	filesWithSymlinkHidden := exampleFileTreeWithSymlinks(t, ".tar_extract_symlink")
	archiveWithSymlinkHiddenPath := filepath.Join(arcDir, "test_with_symlink_hidden.tar")
	_, err = create(ta, filesWithSymlinkHidden, archiveWithSymlinkHiddenPath)
	test.Ok(t, err)

	emptyArchivePath := filepath.Join(arcDir, "empty_test.tar")
	_, err = create(ta, []string{}, emptyArchivePath)
	test.Ok(t, err)

	badArchivePath := filepath.Join(arcDir, "bad_test.tar")
	test.Ok(t, ioutil.WriteFile(badArchivePath, []byte("hello\ndrone\n"), 0644))

	filesAbs := exampleFileTree(t, ".tar_extract_absolute", testAbs)
	archiveAbsPath := filepath.Join(arcDir, "test_absolute.tar")
	_, err = create(ta, filesAbs, archiveAbsPath)
	test.Ok(t, err)

	for _, tc := range []struct {
		name        string
		ta          *Archive
		archivePath string
		srcs        []string
		written     int64
		err         error
	}{
		{
			name:        "non-existing archive",
			ta:          New(log.NewNopLogger(), testRootMounted, false),
			archivePath: "idonotexist",
			srcs:        []string{},
			written:     0,
			err:         os.ErrNotExist,
		},
		{
			name:        "non-existing root destination",
			ta:          New(log.NewNopLogger(), testRootMounted, false),
			archivePath: emptyArchivePath,
			srcs:        []string{},
			written:     0,
			err:         nil,
		},
		{
			name:        "empty archive",
			ta:          New(log.NewNopLogger(), testRootMounted, false),
			archivePath: emptyArchivePath,
			srcs:        []string{},
			written:     0,
			err:         nil,
		},
		{
			name:        "bad archives",
			ta:          New(log.NewNopLogger(), testRootMounted, false),
			archivePath: badArchivePath,
			srcs:        []string{},
			written:     0,
			err:         ErrArchiveNotReadable,
		},
		{
			name:        "existing archive",
			ta:          New(log.NewNopLogger(), testRootMounted, false),
			archivePath: archivePath,
			srcs:        files,
			written:     43,
			err:         nil,
		},
		{
			name:        "existing archive with nested files",
			ta:          New(log.NewNopLogger(), testRootMounted, false),
			archivePath: nestedArchivePath,
			srcs:        nestedFiles,
			written:     56,
			err:         nil,
		},
		{
			name:        "existing archive with symbolic links",
			ta:          New(log.NewNopLogger(), testRootMounted, false),
			archivePath: archiveWithSymlinkPath,
			srcs:        filesWithSymlink,
			written:     43,
			err:         nil,
		},
		{
			name:        "existing archive with hidden symbolic links",
			ta:          New(log.NewNopLogger(), testRootMounted, false),
			archivePath: archiveWithSymlinkHiddenPath,
			srcs:        filesWithSymlinkHidden,
			written:     43,
			err:         nil,
		},
		{
			name:        "absolute mount paths",
			ta:          New(log.NewNopLogger(), testRootMounted, true),
			archivePath: archiveAbsPath,
			srcs:        filesAbs,
			written:     43,
			err:         nil,
		},
	} {
		tc := tc // NOTE: https://github.com/golang/go/wiki/CommonMistakes#using-goroutines-on-loop-iterator-variables
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Setup
			var absSrcs []string
			var relativeSrcs []string

			for _, src := range tc.srcs {
				if strings.HasPrefix(src, "/") {
					absSrcs = append(absSrcs, src)
				} else {
					relativeSrcs = append(relativeSrcs, src)
				}
			}

			dstDir, dstDirClean := test.CreateTempDir(t, "tar_extract_"+tc.name, testRootExtracted)
			t.Cleanup(dstDirClean)

			// Run
			for _, src := range absSrcs {
				test.Ok(t, os.RemoveAll(src))
			}
			written, err := extract(tc.ta, tc.archivePath, dstDir)
			if err != nil {
				test.Expected(t, err, tc.err)
				return
			}

			// Test
			test.Assert(t, written == tc.written, "case %q: written bytes got %d want %v", tc.name, written, tc.written)
			for _, src := range absSrcs {
				test.Exists(t, src)
			}
			test.EqualDirs(t, dstDir, testRootMounted, relativeSrcs)
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

		written, err := a.Create(srcs, pw)
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

func exampleFileTree(t *testing.T, name string, in string) []string {
	file, fileClean := test.CreateTempFile(t, name, []byte("hello\ndrone!\n"), in) // 13 bytes
	t.Cleanup(fileClean)

	dir, dirClean := test.CreateTempFilesInDir(t, name, []byte("hello\ngo!\n"), in) // 10 bytes
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

	dir, dirClean := test.CreateTempFilesInDir(t, name, []byte("hello\ngo!\n"), testRootMounted) // 10 bytes
	t.Cleanup(dirClean)

	symDir, cleanup := test.CreateTempDir(t, name, testRootMounted)
	t.Cleanup(cleanup)

	symlink := filepath.Join(symDir, name+"_symlink.testfile")
	test.Ok(t, os.Symlink(file, symlink))
	t.Cleanup(func() { os.Remove(symlink) })

	return []string{file, dir, symDir}
}
