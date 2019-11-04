package cache

import (
	"compress/flate"
)

const (
	DefaultCompressionLevel = flate.DefaultCompression
	DefaultArchiveFormat    = "tar"
)

type options struct {
	archiveFmt       string
	compressionLevel int
	skipSymlinks     bool
}

// Option overrides behavior of Cache.
type Option interface {
	apply(*options)
}

type optionFunc func(*options)

func (f optionFunc) apply(o *options) {
	f(o)
}

// WithSkipSymlinks sets skip symlink option.
func WithSkipSymlinks(b bool) Option {
	return optionFunc(func(o *options) {
		o.skipSymlinks = b
	})
}

// WithArchiveFormat sets archive format option.
func WithArchiveFormat(s string) Option {
	return optionFunc(func(o *options) {
		o.archiveFmt = s
	})
}

// WithCompressionLevel sets compression level option.
func WithCompressionLevel(i int) Option {
	return optionFunc(func(o *options) {
		o.compressionLevel = i
	})
}
