package zstd

import (
	"fmt"
	"io"

	"github.com/go-kit/log"
	"github.com/klauspost/compress/zstd"
	"github.com/meltwater/drone-cache/archive/tar"
	"github.com/meltwater/drone-cache/internal"
)

// Archive implements archive for zstd.
type Archive struct {
	logger log.Logger

	root             string
	compressionLevel int
	skipSymlinks     bool
}

// New creates an archive that uses the .tar.zst file format.
func New(logger log.Logger, root string, skipSymlinks bool, compressionLevel int) *Archive {
	return &Archive{logger, root, compressionLevel, skipSymlinks}
}

// Create writes content of the given source to an archive, returns written bytes.
func (a *Archive) Create(srcs []string, w io.Writer, isRelativePath bool) (int64, error) {
	level := zstd.SpeedDefault
	if a.compressionLevel != -1 {
		level = zstd.EncoderLevelFromZstd(a.compressionLevel)
	}
	zw, err := zstd.NewWriter(w, zstd.WithEncoderLevel(level))
	if err != nil {
		return 0, fmt.Errorf("zstd create archive writer, %w", err)
	}

	defer internal.CloseWithErrLogf(a.logger, zw, "zstd writer")

	wBytes, err := tar.New(a.logger, a.root, a.skipSymlinks).Create(srcs, zw, isRelativePath)
	if err != nil {
		return 0, fmt.Errorf("zstd create archive, %w", err)
	}

	return wBytes, nil
}

// Extract reads content from the given archive reader and restores it to the destination, returns written bytes.
func (a *Archive) Extract(dst string, r io.Reader) (int64, error) {
	zr, err := zstd.NewReader(r)
	if err != nil {
		return 0, fmt.Errorf("zstd create extract archive reader, %w", err)
	}

	defer internal.CloseWithErrLogf(a.logger, zr.IOReadCloser(), "zstd reader")

	eBytes, err := tar.New(a.logger, a.root, a.skipSymlinks).Extract(dst, zr)
	if err != nil {
		return 0, fmt.Errorf("zstd extract archive, %w", err)
	}

	return eBytes, nil
}
