package tar

import (
	"archive/tar"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-kit/kit/log"

	"github.com/meltwater/drone-cache/internal"
)

const defaultDirPermission = 0755

var (
	// ErrSourceNotReachable means that given source is not reachable.
	ErrSourceNotReachable = errors.New("source not reachable")
	// ErrArchiveNotReadable means that given archive not readable/corrupted.
	ErrArchiveNotReadable = errors.New("archive not readable")
)

// Archive implements archive for tar.
type Archive struct {
	logger log.Logger

	root         string
	skipSymlinks bool
}

// New creates an archive that uses the .tar file format.
func New(logger log.Logger, root string, skipSymlinks bool) *Archive {
	return &Archive{logger, root, skipSymlinks}
}

// Create writes content of the given source to an archive, returns written bytes.
func (a *Archive) Create(srcs []string, w io.Writer) (int64, error) {
	tw := tar.NewWriter(w)
	defer internal.CloseWithErrLogf(a.logger, tw, "tar writer")

	var written int64

	for _, src := range srcs {
		_, err := os.Lstat(src)
		if err != nil {
			return written, fmt.Errorf("make sure file or directory readable <%s>: %v,, %w", src, err, ErrSourceNotReachable)
		}

		if err := filepath.Walk(src, writeToArchive(tw, a.root, a.skipSymlinks, &written)); err != nil {
			return written, fmt.Errorf("walk, add all files to archive, %w", err)
		}
	}

	return written, nil
}

// nolint: lll
func writeToArchive(tw *tar.Writer, root string, skipSymlinks bool, written *int64) func(string, os.FileInfo, error) error {
	return func(path string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if fi == nil {
			return errors.New("no file info")
		}

		// Create header for Regular files and Directories
		h, err := tar.FileInfoHeader(fi, fi.Name())
		if err != nil {
			return fmt.Errorf("create header for <%s>, %w", path, err)
		}

		if fi.Mode()&os.ModeSymlink != 0 { // isSymbolic
			if skipSymlinks {
				return nil
			}

			var err error
			if h, err = createSymlinkHeader(fi, path); err != nil {
				return fmt.Errorf("create header for symbolic link, %w", err)
			}
		}

		var name string

		if strings.HasPrefix(path, "/") {
			name, err = filepath.Abs(path)
		} else {
			name, err = relative(root, path)
		}

		if err != nil {
			return fmt.Errorf("relative name <%s>: <%s>, %w", path, root, err)
		}

		h.Name = name

		if err := tw.WriteHeader(h); err != nil {
			return fmt.Errorf("write header for <%s>, %w", path, err)
		}

		if !fi.Mode().IsRegular() {
			return nil
		}

		n, err := writeFileToArchive(tw, path)
		if err != nil {
			return fmt.Errorf("write file to archive, %w", err)
		}

		*written += n
		// Alternatives:
		// *written += h.FileInfo().Size()
		// *written += fi.Size()

		return nil
	}
}

func relative(parent string, path string) (string, error) {
	name := filepath.Base(path)

	rel, err := filepath.Rel(parent, filepath.Dir(path))
	if err != nil {
		return "", fmt.Errorf("relative path <%s>, base <%s>, %w", rel, name, err)
	}

	// NOTICE: filepath.Rel puts "../" when given path is not under parent.
	for strings.HasPrefix(rel, "../") {
		rel = strings.TrimPrefix(rel, "../")
	}

	rel = filepath.ToSlash(rel)

	return strings.TrimPrefix(filepath.Join(rel, name), "/"), nil
}

func createSymlinkHeader(fi os.FileInfo, path string) (*tar.Header, error) {
	lnk, err := os.Readlink(path)
	if err != nil {
		return nil, fmt.Errorf("read link <%s>, %w", path, err)
	}

	h, err := tar.FileInfoHeader(fi, lnk)
	if err != nil {
		return nil, fmt.Errorf("create symlink header for <%s>, %w", path, err)
	}

	return h, nil
}

func writeFileToArchive(tw io.Writer, path string) (n int64, err error) {
	f, err := os.Open(path)
	if err != nil {
		return 0, fmt.Errorf("open file <%s>, %w", path, err)
	}

	defer internal.CloseWithErrCapturef(&err, f, "write file to archive <%s>", path)

	written, err := io.Copy(tw, f)
	if err != nil {
		return written, fmt.Errorf("copy the file <%s> data to the tarball, %w", path, err)
	}

	return written, nil
}

// Extract reads content from the given archive reader and restores it to the destination, returns written bytes.
func (a *Archive) Extract(dst string, r io.Reader) (int64, error) {
	var (
		written int64
		tr      = tar.NewReader(r)
	)

	for {
		h, err := tr.Next()

		switch {
		case err == io.EOF: // if no more files are found return
			return written, nil
		case err != nil: // return any other error
			return written, fmt.Errorf("tar reader <%v>, %w", err, ErrArchiveNotReadable)
		case h == nil: // if the header is nil, skip it
			continue
		}

		var target string
		if dst == h.Name || strings.HasPrefix(h.Name, "/") {
			target = h.Name
		} else {
			name, err := relative(dst, h.Name)
			if err != nil {
				return 0, fmt.Errorf("relative name, %w", err)
			}

			target = filepath.Join(dst, name)
		}

		if err := os.MkdirAll(filepath.Dir(target), defaultDirPermission); err != nil {
			return 0, fmt.Errorf("ensure directory <%s>, %w", target, err)
		}

		switch h.Typeflag {
		case tar.TypeDir:
			if err := extractDir(h, target); err != nil {
				return written, err
			}

			continue
		case tar.TypeReg, tar.TypeRegA, tar.TypeChar, tar.TypeBlock, tar.TypeFifo:
			n, err := extractRegular(h, tr, target)
			written += n

			if err != nil {
				return written, fmt.Errorf("extract regular file, %w", err)
			}

			continue
		case tar.TypeSymlink:
			if err := extractSymlink(h, target); err != nil {
				return written, fmt.Errorf("extract symbolic link, %w", err)
			}

			continue
		case tar.TypeLink:
			if err := extractLink(h, target); err != nil {
				return written, fmt.Errorf("extract link, %w", err)
			}

			continue
		case tar.TypeXGlobalHeader:
			continue
		default:
			return written, fmt.Errorf("extract %s, unknown type flag: %c", target, h.Typeflag)
		}
	}
}

func extractDir(h *tar.Header, target string) error {
	if err := os.MkdirAll(target, os.FileMode(h.Mode)); err != nil {
		return fmt.Errorf("create directory <%s>, %w", target, err)
	}

	return nil
}

func extractRegular(h *tar.Header, tr io.Reader, target string) (n int64, err error) {
	f, err := os.OpenFile(target, os.O_CREATE|os.O_RDWR, os.FileMode(h.Mode))
	if err != nil {
		return 0, fmt.Errorf("open extracted file for writing <%s>, %w", target, err)
	}

	defer internal.CloseWithErrCapturef(&err, f, "extract regular <%s>", target)

	written, err := io.Copy(f, tr)
	if err != nil {
		return written, fmt.Errorf("copy extracted file for writing <%s>, %w", target, err)
	}

	return written, nil
}

func extractSymlink(h *tar.Header, target string) error {
	if err := unlink(target); err != nil {
		return fmt.Errorf("unlink <%s>, %w", target, err)
	}

	if err := os.Symlink(h.Linkname, target); err != nil {
		return fmt.Errorf("create symbolic link <%s>, %w", target, err)
	}

	return nil
}

func extractLink(h *tar.Header, target string) error {
	if err := unlink(target); err != nil {
		return fmt.Errorf("unlink <%s>, %w", target, err)
	}

	if err := os.Link(h.Linkname, target); err != nil {
		return fmt.Errorf("create hard link <%s>, %w", h.Linkname, err)
	}

	return nil
}

func unlink(path string) error {
	_, err := os.Lstat(path)
	if err == nil {
		return os.Remove(path)
	}

	return nil
}
