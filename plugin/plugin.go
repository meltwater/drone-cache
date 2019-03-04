// Package plugin for caching directories using given backends
package plugin

import (
	"fmt"
	"log"
	"path/filepath"
	"time"

	"github.com/pkg/errors"

	"github.com/meltwater/drone-cache/cache"
	"github.com/meltwater/drone-cache/cache/backend"
	"github.com/meltwater/drone-cache/metadata"
	"github.com/meltwater/drone-cache/plugin/cachekey"
)

type (
	// Config plugin-specific parameters and secrets
	Config struct {
		ArchiveFormat string
		Backend       string
		CacheKey      string

		Debug   bool
		Rebuild bool
		Restore bool

		Mount []string

		S3         backend.S3Config
		FileSystem backend.FileSystemConfig
	}

	// Plugin stores metadata about current plugin
	Plugin struct {
		Metadata metadata.Metadata
		Config   Config
	}

	// Error recognized error from plugin
	Error string
)

func (e Error) Error() string { return string(e) }

// Exec entry point of Plugin, where the magic happens
func (p *Plugin) Exec() error {
	c := p.Config

	// 1. Check parameters
	if c.Debug {
		log.Println("DEBUG MODE enabled!")
		log.Printf("Plugin initialized with config: %+v", p.Config)
		log.Printf("Plugin initialized with metadata: %+v", p.Metadata)
	}

	if c.Rebuild && c.Restore {
		return errors.New("rebuild and restore are mutually exclusive, please set only one of them")
	}

	_, err := cachekey.ParseTemplate(c.CacheKey)
	if err != nil {
		msg := fmt.Sprintf("could not parse <%s> as cache key template, falling back to default", c.CacheKey)
		return errors.Wrap(err, msg)
	}

	// 2. Initialize backend
	backend, err := initializeBackend(c)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("could not initialize <%s> as backend", c.Backend))
	}

	// 3. Initialize cache
	cch := cache.New(backend, c.ArchiveFormat)

	// 4. Select mode
	if c.Rebuild {
		if err := processRebuild(cch, p.Config.CacheKey, p.Config.Mount, p.Metadata); err != nil {
			return Error(fmt.Sprintf("WARNING: could not build cache. process rebuild failed, %v\n", err))
		}
	}

	if c.Restore {
		if err := processRestore(cch, p.Config.CacheKey, p.Config.Mount, p.Metadata); err != nil {
			return Error(fmt.Sprintf("WARNING: could not restore cache. process restore failed, %v\n", err))
		}
	}

	return nil
}

// initializeBackend initializes backend using given configuration
func initializeBackend(c Config) (cache.Backend, error) {
	switch c.Backend {
	case "s3":
		log.Println("IMPORTANT: using aws s3 as backend")
		return backend.InitializeS3Backend(c.S3, c.Debug)
	case "filesystem":
		log.Println("IMPORTANT: using filesystem as backend")
		return backend.InitializeFileSystemBackend(c.FileSystem, c.Debug)
	default:
		return nil, errors.New("unknown backend")
	}
}

// processRebuild the remote cache from the local environment
func processRebuild(c cache.Cache, cacheKeyTmpl string, mountedDirs []string, m metadata.Metadata) error {
	now := time.Now()
	branch := m.Commit.Branch

	for _, mount := range mountedDirs {
		key, err := cacheKey(m, cacheKeyTmpl, mount, branch)
		if err != nil {
			return errors.Wrap(err, "could not generate cache key")
		}
		path := filepath.Join(m.Repo.Name, key)

		log.Printf("rebuilding cache for directory <%s> to remote cache <%s>", mount, path)
		if err := c.Push(mount, path); err != nil {
			return errors.Wrap(err, "could not upload")
		}
	}
	log.Printf("cache built in %v", time.Since(now))
	return nil
}

// processRestore the local environment from the remote cache
func processRestore(c cache.Cache, cacheKeyTmpl string, mountedDirs []string, m metadata.Metadata) error {
	now := time.Now()
	branch := m.Commit.Branch

	for _, mount := range mountedDirs {
		key, err := cacheKey(m, cacheKeyTmpl, mount, branch)
		if err != nil {
			return errors.Wrap(err, "could not generate cache key")
		}
		path := filepath.Join(m.Repo.Name, key)

		log.Printf("restoring directory <%s> from remote cache <%s>", mount, path)
		if err := c.Pull(path, mount); err != nil {
			return errors.Wrap(err, "could not download")
		}
	}
	log.Printf("cache restored in %v", time.Since(now))
	return nil
}

// Helpers

// cacheKey generates key from given template as parameter or fallbacks hash
func cacheKey(p metadata.Metadata, cacheKeyTmpl, mount, branch string) (string, error) {
	log.Println("using provided cache key template")
	key, err := cachekey.Generate(cacheKeyTmpl, mount, metadata.Metadata{
		Build:  p.Build,
		Commit: p.Commit,
		Repo:   p.Repo,
	})

	if err != nil {
		log.Printf("%v, falling back to default key", err)
		key, err = cachekey.Hash(mount, branch)
		if err != nil {
			return "", errors.Wrap(err, "could not generate hash key for mounted")
		}
	}

	return key, nil
}
