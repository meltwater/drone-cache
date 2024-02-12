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

// Rebuilder is an interface represents a rebuild action.
type Rebuilder interface {
	// Rebuild rebuilds cache from the files provided with given paths.
	Rebuild(srcs []string) error
}

// Restorer is an interface represents a restore action.
type Restorer interface {
	// Restore restores files from the cache provided with given paths.
	Restore(srcs []string) error
}

// Flusher is an interface represents a flush action.
type Flusher interface {
	// Flush removes files from the cache using given paths.
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
		NewRebuilder(log.With(logger, "component", "rebuilder"), s, a, g,
			options.fallbackGenerator, options.namespace, options.override, options.gracefulDetect),
		NewRestorer(log.With(logger, "component", "restorer"), s, a, g,
			options.fallbackGenerator, options.namespace, options.failRestoreIfKeyNotPresent, options.disableCacheKeySeparator),
		NewFlusher(log.With(logger, "component", "flusher"), s, time.Hour),
	}
}
