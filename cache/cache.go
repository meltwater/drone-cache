// Package cache provides functionality for cache storage
package cache

import (
	"time"

	"github.com/go-kit/kit/log"
	"github.com/meltwater/drone-cache/archive"
	"github.com/meltwater/drone-cache/key"
	"github.com/meltwater/drone-cache/storage"
)

// Cache defines Cache functionality and stores configuration.
type Cache interface {
	Rebuilder
	Restorer
	Flusher
}

// Rebuilder TODO
type Rebuilder interface {
	// Rebuild TODO
	Rebuild(srcs []string) error
}

// Restorer TODO
type Restorer interface {
	// Restore TODO
	Restore(srcs []string) error
}

// Flusher TODO
type Flusher interface {
	// Flush TODO
	Flush(srcs []string) error
}

type cache struct {
	Rebuilder
	Restorer
	Flusher
}

// New creates a new cache with given parameters.
func New(logger log.Logger, s storage.Storage, a archive.Archive, g key.Generator, opts ...Option) Cache {
	options := options{}

	for _, o := range opts {
		o.apply(&options)
	}

	return &cache{
		NewRebuilder(log.With(logger, "component", "rebuilder"), s, a, g, options.fallbackGenerator, options.namespace, options.override),
		NewRestorer(log.With(logger, "component", "restorer"), s, a, g, options.fallbackGenerator, options.namespace),
		NewFlusher(log.With(logger, "component", "flusher"), s, time.Hour),
	}
}
