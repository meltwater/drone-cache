package cache

import "github.com/meltwater/drone-cache/key"

type options struct {
	namespace                  string
	fallbackGenerator          key.Generator
	override                   bool
	failRestoreIfKeyNotPresent bool
	gracefulDetect             bool
	enableCacheKeySeparator    bool
}

// Option overrides behavior of Archive.
type Option interface {
	apply(*options)
}

type optionFunc func(*options)

func (f optionFunc) apply(o *options) {
	f(o)
}

// WithNamespace sets namespace option.
func WithNamespace(s string) Option {
	return optionFunc(func(o *options) {
		o.namespace = s
	})
}

// WithFallbackGenerator sets fallback key generator option.
func WithFallbackGenerator(g key.Generator) Option {
	return optionFunc(func(o *options) {
		o.fallbackGenerator = g
	})
}

// WithOverride sets object should be overriten even if it already exists.
func WithOverride(override bool) Option {
	return optionFunc(func(o *options) {
		o.override = override
	})
}

// WithGracefulDetect sets option to fail sve if directory does not exist.
func WithGracefulDetect(gracefulDetect bool) Option {
	return optionFunc(func(o *options) {
		o.gracefulDetect = gracefulDetect
	})
}

// WithFailRestoreIfKeyNotPresent sets option to fail restore if key does not exist.
func WithFailRestoreIfKeyNotPresent(b bool) Option {
	return optionFunc(func(o *options) {
		o.failRestoreIfKeyNotPresent = b
	})
}

// WithEnableCacheKeySeparator enables addition of '/' to the specified cache key.
func WithEnableCacheKeySeparator(b bool) Option {
	return optionFunc(func(o *options) {
		o.enableCacheKeySeparator = b
	})
}
