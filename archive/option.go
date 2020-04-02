package archive

type options struct {
	compressionLevel int
	skipSymlinks     bool
}

// Option overrides behavior of Archive.
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

// WithCompressionLevel sets compression level option.
func WithCompressionLevel(i int) Option {
	return optionFunc(func(o *options) {
		o.compressionLevel = i
	})
}
