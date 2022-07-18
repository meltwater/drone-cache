package gzip

import (
	"compress/gzip"
	"fmt"
	"io"

	"github.com/go-kit/log"
	"github.com/meltwater/drone-cache/archive/tar"
	"github.com/meltwater/drone-cache/internal"
)

// Archive implements archive for gzip.
type Archive struct {
	logger log.Logger

	root             string
	compressionLevel int
	skipSymlinks     bool
}

// New creates an archive that uses the .tar.gz file format.
func New(logger log.Logger, root string, skipSymlinks bool, compressionLevel int) *Archive {
	return &Archive{logger, root, compressionLevel, skipSymlinks}
}

// Create writes content of the given source to an archive, returns written bytes.
func (a *Archive) Create(srcs []string, w io.Writer) (int64, error) {
	gw, err := gzip.NewWriterLevel(w, a.compressionLevel)
	if err != nil {
		return 0, fmt.Errorf("create archive writer, %w", err)
	}

	defer internal.CloseWithErrLogf(a.logger, gw, "gzip writer")

	wBytes, err := tar.New(a.logger, a.root, a.skipSymlinks).Create(srcs, gw)
	if err != nil {
		return 0, fmt.Errorf("writing create archive bytes: %w", err)
	}

	return wBytes, nil
}

// Extract reads content from the given archive reader and restores it to the destination, returns written bytes.
func (a *Archive) Extract(dst string, r io.Reader) (int64, error) {
	gr, err := gzip.NewReader(r)
	if err != nil {
		return 0, fmt.Errorf("create archive extractor: %w", err)
	}

	defer internal.CloseWithErrLogf(a.logger, gr, "gzip reader")

	eBytes, err := tar.New(a.logger, a.root, a.skipSymlinks).Extract(dst, gr)
	if err != nil {
		return 0, fmt.Errorf("extracting archive bytes: %w", err)
	}

	return eBytes, nil
}
