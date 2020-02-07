// Package plugin for caching directories using given backends
package plugin

import (
	"errors"
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

		S3           backend.S3Config
		FileSystem   backend.FileSystemConfig
		SFTP         backend.SFTPConfig
		Azure        backend.AzureConfig
		CloudStorage backend.CloudStorageConfig
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
		return fmt.Errorf("parse, <%s> as cache key template, falling back to default %w", c.CacheKey, err)
	}

	// 2. Initialize backend
	backend, err := initializeBackend(p.Logger, c)
	if err != nil {
		return fmt.Errorf("initialize, <%s> as backend %w", c.Backend, err)
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
			return Error(fmt.Sprintf("[WARNING] build cache, process rebuild failed, %v\n", err))
		}
	}

	if c.Restore {
		if err := processRestore(p.Logger, cch, p.Config.CacheKey, p.Config.Mount, p.Metadata); err != nil {
			return Error(fmt.Sprintf("[WARNING] restore cache, process restore failed, %v\n", err))
		}
	}

	return nil
}

// initializeBackend initializes backend using given configuration
func initializeBackend(logger log.Logger, c Config) (cache.Backend, error) {
	switch c.Backend {
	case "azure":
		level.Warn(logger).Log("msg", "using azure blob as backend")
		return backend.InitializeAzureBackend(logger, c.Azure, c.Debug)
	case "s3":
		level.Warn(logger).Log("msg", "using aws s3 as backend")
		return backend.InitializeS3Backend(logger, c.S3, c.Debug)
	case "cloudstorage":
		level.Warn(logger).Log("msg", "using gc storage as backend")
		return backend.InitializeGCSBackend(logger, c.CloudStorage, c.Debug)
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
func processRebuild(l log.Logger, c cache.Cache, cacheKeyTmpl string, mountedDirs []string, m metadata.Metadata) error {
	now := time.Now()
	branch := m.Commit.Branch

	for _, mount := range mountedDirs {
		if _, err := os.Stat(mount); err != nil {
			return fmt.Errorf("mount <%s>, make sure file or directory exists and readable %w", mount, err)
		}

		key, err := cacheKey(l, m, cacheKeyTmpl, mount, branch)
		if err != nil {
			return fmt.Errorf("generate cache key %w", err)
		}

		path := filepath.Join(m.Repo.Name, key)

		level.Info(l).Log("msg", "rebuilding cache for directory", "local", mount, "remote", path)

		if err := c.Push(mount, path); err != nil {
			return fmt.Errorf("upload %w", err)
		}
	}

	level.Info(l).Log("msg", "cache built", "took", time.Since(now))

	return nil
}

// processRestore the local environment from the remote cache
func processRestore(l log.Logger, c cache.Cache, cacheKeyTmpl string, mountedDirs []string, m metadata.Metadata) error {
	now := time.Now()
	branch := m.Commit.Branch

	for _, mount := range mountedDirs {
		key, err := cacheKey(l, m, cacheKeyTmpl, mount, branch)
		if err != nil {
			return fmt.Errorf("generate cache key %w", err)
		}

		path := filepath.Join(m.Repo.Name, key)
		level.Info(l).Log("msg", "restoring directory", "local", mount, "remote", path)

		if err := c.Pull(path, mount); err != nil {
			return fmt.Errorf("download %w", err)
		}
	}

	level.Info(l).Log("msg", "cache restored", "took", time.Since(now))

	return nil
}

// Helpers

// cacheKey generates key from given template as parameter or fallbacks hash
func cacheKey(l log.Logger, p metadata.Metadata, cacheKeyTmpl, mount, branch string) (string, error) {
	level.Info(l).Log("msg", "using provided cache key template")

	key, err := cachekey.Generate(cacheKeyTmpl, mount, metadata.Metadata{
		Build:  p.Build,
		Commit: p.Commit,
		Repo:   p.Repo,
	})

	if err != nil {
		level.Error(l).Log("msg", "falling back to default key", "err", err)
		key, err = cachekey.Hash(mount, branch)

		if err != nil {
			return "", fmt.Errorf("generate hash key for mounted %w", err)
		}
	}

	return key, nil
}
