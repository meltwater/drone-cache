package cache

import "github.com/meltwater/drone-cache/key"

type options struct {
	namespace         string
	fallbackGenerator key.Generator
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
