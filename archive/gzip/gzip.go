package gzip

import (
	"compress/gzip"
	"fmt"
	"io"

	"github.com/meltwater/drone-cache/archive/tar"
	"github.com/meltwater/drone-cache/internal"

	"github.com/go-kit/kit/log"
)

// pArchive TODO
type Archive struct {
	logger log.Logger

	compressionLevel int
	skipSymlinks     bool
}

// New creates an archive that uses the .tar.gz file format.
func New(logger log.Logger, skipSymlinks bool, compressionLevel int) *Archive {
	return &Archive{logger, compressionLevel, skipSymlinks}
}

// Create writes content of the given source to an archive, returns written bytes.
func (a *Archive) Create(srcs []string, w io.Writer) (int64, error) {
	gw, err := gzip.NewWriterLevel(w, a.compressionLevel)
	if err != nil {
		return 0, fmt.Errorf("create archive writer %w", err)
	}

	defer internal.CloseWithErrLogf(a.logger, gw, "gzip writer")

	return tar.New(a.logger, a.skipSymlinks).Create(srcs, gw)
}

// Extract reads content from the given archive reader and restores it to the destination, returns written bytes.
func (a *Archive) Extract(dst string, r io.Reader) (int64, error) {
	gr, err := gzip.NewReader(r)
	if err != nil {
		return 0, err
	}

	defer internal.CloseWithErrLogf(a.logger, gr, "gzip reader")

	return tar.New(a.logger, a.skipSymlinks).Extract(dst, gr)
}
