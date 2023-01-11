package archive

import (
	"compress/flate"
	"io"

	"github.com/meltwater/drone-cache/archive/gzip"
	"github.com/meltwater/drone-cache/archive/tar"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
)

const (
	Gzip = "gzip"
	Tar  = "tar"

	DefaultCompressionLevel = flate.DefaultCompression
	DefaultArchiveFormat    = Tar
)

// Archive is an interface that defines exposed behavior of archive formats.
type Archive interface {
	// Create writes content of the given source to an archive, returns written bytes.
	// Similar to io.WriterTo.
	// If isRelativePath is true, it clones using the path, else it clones using a path
	// combining archive's root with the path.
	Create(srcs []string, w io.Writer, isRelativePath bool) (int64, error)

	// Extract reads content from the given archive reader and restores it to the destination, returns written bytes.
	// Similar to io.ReaderFrom.
	Extract(dst string, r io.Reader) (int64, error)
}

// FromFormat determines which archive to use from given archive format.
func FromFormat(logger log.Logger, root string, format string, opts ...Option) Archive {
	options := options{
		compressionLevel: DefaultCompressionLevel,
	}

	for _, o := range opts {
		o.apply(&options)
	}

	switch format {
	case Gzip:
		return gzip.New(logger, root, options.skipSymlinks, options.compressionLevel)
	case Tar:
		return tar.New(logger, root, options.skipSymlinks)
	default:
		level.Error(logger).Log("msg", "unknown archive format", "format", format)
		return tar.New(logger, root, options.skipSymlinks) // DefaultArchiveFormat
	}
}
