// Package plugin for caching directories using given backends
package plugin

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/meltwater/drone-cache/cache"
	"github.com/meltwater/drone-cache/cache/backend"
	"github.com/meltwater/drone-cache/metadata"
	"github.com/meltwater/drone-cache/plugin/cachekey"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/pkg/errors"
)

type (
	// Config plugin-specific parameters and secrets.
	Config struct {
		ArchiveFormat string
		Backend       string
		CacheKey      string

		CompressionLevel int

		Debug        bool
		SkipSymlinks bool
		Rebuild      bool
		Restore      bool

		Mount []string

		S3         backend.S3Config
		FileSystem backend.FileSystemConfig
		SFTP       backend.SFTPConfig
	}

	// Plugin stores metadata about current plugin.
	Plugin struct {
		Logger   log.Logger
		Metadata metadata.Metadata
		Config   Config
	}

	// Error recognized error from plugin.
	Error string
)

func (e Error) Error() string { return string(e) }

// Exec entry point of Plugin, where the magic happens.
func (p *Plugin) Exec() error {
	c := p.Config

	// 1. Check parameters
	if c.Debug {
		level.Debug(p.Logger).Log("msg", "DEBUG MODE enabled!")

		for _, pair := range os.Environ() {
			level.Debug(p.Logger).Log("var", pair)
		}

		level.Debug(p.Logger).Log("msg", "plugin initialized with config", "config", fmt.Sprintf("%+v", p.Config))
		level.Debug(p.Logger).Log("msg", "plugin initialized with metadata", "metadata", fmt.Sprintf("%+v", p.Metadata))
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
	backend, err := initializeBackend(p.Logger, c)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("could not initialize <%s> as backend", c.Backend))
	}

	// 3. Initialize cache
	cch := cache.New(p.Logger, backend,
		cache.WithArchiveFormat(c.ArchiveFormat),
		cache.WithSkipSymlinks(c.SkipSymlinks),
		cache.WithCompressionLevel(c.CompressionLevel),
	)

	// 4. Select mode
	if c.Rebuild {
		if err := processRebuild(p.Logger, cch, p.Config.CacheKey, p.Config.Mount, p.Metadata); err != nil {
			// TODO: !!! new errors
			return Error(fmt.Sprintf("[WARNING] could not build cache, process rebuild failed, %v\n", err))
		}
	}

	if c.Restore {
		if err := processRestore(p.Logger, cch, p.Config.CacheKey, p.Config.Mount, p.Metadata); err != nil {
			// TODO: !!! new errors
			return Error(fmt.Sprintf("[WARNING] could not restore cache, process restore failed, %v\n", err))
		}
	}

	return nil
}

// initializeBackend initializes backend using given configuration
func initializeBackend(logger log.Logger, c Config) (cache.Backend, error) {
	switch c.Backend {
	case "s3":
		level.Warn(logger).Log("msg", "using aws s3 as backend")
		return backend.InitializeS3Backend(logger, c.S3, c.Debug)
	case "filesystem":
		level.Warn(logger).Log("msg", "using filesystem as backend")
		return backend.InitializeFileSystemBackend(logger, c.FileSystem, c.Debug)
	case "sftp":
		level.Warn(logger).Log("msg", "using sftp as backend")
		return backend.InitializeSFTPBackend(logger, c.SFTP, c.Debug)
	default:
		return nil, errors.New("unknown backend")
	}
}

// processRebuild the remote cache from the local environment
func processRebuild(logger log.Logger, c cache.Cache, cacheKeyTmpl string, mountedDirs []string, m metadata.Metadata) error {
	now := time.Now()
	branch := m.Commit.Branch

	for _, mount := range mountedDirs {
		if _, err := os.Stat(mount); err != nil {
			return errors.Wrap(err, fmt.Sprintf("could not mount <%s>, make sure file or directory exists and readable", mount))
		}

		key, err := cacheKey(logger, m, cacheKeyTmpl, mount, branch)
		if err != nil {
			return errors.Wrap(err, "could not generate cache key")
		}

		path := filepath.Join(m.Repo.Name, key)

		level.Info(logger).Log("msg", "rebuilding cache for directory", "local", mount, "remote", path)

		if err := c.Push(mount, path); err != nil {
			return errors.Wrap(err, "could not upload")
		}
	}

	level.Info(logger).Log("msg", "cache built", "took", time.Since(now))

	return nil
}

// processRestore the local environment from the remote cache
func processRestore(logger log.Logger, c cache.Cache, cacheKeyTmpl string, mountedDirs []string, m metadata.Metadata) error {
	now := time.Now()
	branch := m.Commit.Branch

	for _, mount := range mountedDirs {
		key, err := cacheKey(logger, m, cacheKeyTmpl, mount, branch)
		if err != nil {
			return errors.Wrap(err, "could not generate cache key")
		}

		path := filepath.Join(m.Repo.Name, key)
		level.Info(logger).Log("msg", "restoring directory", "local", mount, "remote", path)

		if err := c.Pull(path, mount); err != nil {
			return errors.Wrap(err, "could not download")
		}
	}

	level.Info(logger).Log("msg", "cache restored", "took", time.Since(now))

	return nil
}

// Helpers

// cacheKey generates key from given template as parameter or fallbacks hash
func cacheKey(logger log.Logger, p metadata.Metadata, cacheKeyTmpl, mount, branch string) (string, error) {
	level.Info(logger).Log("msg", "using provided cache key template")

	key, err := cachekey.Generate(cacheKeyTmpl, mount, metadata.Metadata{
		Build:  p.Build,
		Commit: p.Commit,
		Repo:   p.Repo,
	})

	if err != nil {
		level.Error(logger).Log("msg", "falling back to default key", "err", err)
		key, err = cachekey.Hash(mount, branch)

		if err != nil {
			return "", errors.Wrap(err, "could not generate hash key for mounted")
		}
	}

	return key, nil
}
