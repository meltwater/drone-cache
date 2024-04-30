// Package plugin for caching directories using given backends
package plugin

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/meltwater/drone-cache/internal/plugin/autodetect"

	"github.com/meltwater/drone-cache/archive"
	"github.com/meltwater/drone-cache/cache"
	"github.com/meltwater/drone-cache/internal/metadata"
	"github.com/meltwater/drone-cache/key"
	keygen "github.com/meltwater/drone-cache/key/generator"
	"github.com/meltwater/drone-cache/storage"
	"github.com/meltwater/drone-cache/storage/backend"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
)

// Error recognized error from plugin.
type Error string

// Error is a sentinel plugin error.
func (e Error) Error() string { return string(e) }

// Unwrap unwraps underlying error.
func (e Error) Unwrap() error { return e }

// Plugin stores metadata about current plugin.
type Plugin struct {
	logger log.Logger

	Metadata metadata.Metadata
	Config   Config
}

// New creates a new plugin.
func New(logger log.Logger) *Plugin {
	return &Plugin{logger: logger}
}

// Exec entry point of Plugin, where the magic happens.
func (p *Plugin) Exec() error { // nolint:funlen
	cfg := p.Config

	// 1. Check parameters
	if cfg.Debug {
		level.Debug(p.logger).Log("msg", "DEBUG MODE enabled!")

		for _, pair := range os.Environ() {
			level.Debug(p.logger).Log("var", pair)
		}

		level.Debug(p.logger).Log("msg", "plugin initialized wth config", "config", fmt.Sprintf("%#v", p.Config))
		level.Debug(p.logger).Log("msg", "plugin initialized with metadata", "metadata", fmt.Sprintf("%#v", p.Metadata))
	}

	// FLUSH

	if cfg.Rebuild && cfg.Restore {
		return errors.New("rebuild and restore are mutually exclusive, please set only one of them")
	}

	var localRoot string
	if p.Config.LocalRoot != "" {
		localRoot = filepath.Clean(p.Config.LocalRoot)
	} else {
		workspace, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("get working directory, %w", err)
		}

		localRoot = workspace
	}

	var options []cache.Option
	// Removing below namespace from cache path so that saved cache path is same for PR and manual runs.
	// if p.Config.RemoteRoot != "" {
	// 	options = append(options, cache.WithNamespace(p.Config.RemoteRoot))
	// } else {
	// 	options = append(options, cache.WithNamespace(p.Metadata.Repo.Name))
	// }

	var generator key.Generator

	switch {
	case cfg.AutoDetect:
		{
			pathOverridden := len(p.Config.Mount) > 0
			dirs, buildTools, cacheKey, err := autodetect.DetectDirectoriesToCache(pathOverridden)
			if err != nil {
				return fmt.Errorf("autodetect enabled but failed to detect, falling back to default, %w", err)
			}
			if len(buildTools) > 0 {
				p.logger.Log("msg", "build tools detected: "+strings.Join(buildTools, ", ")) //nolint: errcheck
			} else if pathOverridden {
				p.logger.Log("msg", "using provided cache path") //nolint: errcheck
			} else {
				p.logger.Log("msg", "no supported build tool detected") //nolint: errcheck
			}
			if !pathOverridden {
				p.Config.Mount = dirs
				options = append(options, cache.WithGracefulDetect(true))
			} else {
				options = append(options, cache.WithGracefulDetect(false))
			}
			if cfg.CacheKeyTemplate != "" {
				cacheKey = cfg.CacheKeyTemplate
			} else if cacheKey == "" {
				cacheKey = "default"
			}
			generator = keygen.NewMetadata(p.logger, cfg.AccountID+"/"+cacheKey, p.Metadata)
			if err := generator.Check(); err != nil {
				return fmt.Errorf("parse failed, falling back to default, %w", err)
			}

			options = append(options, cache.WithFallbackGenerator(keygen.NewHash(cfg.AccountID+p.Metadata.Commit.Branch)))
		}
	case cfg.CacheKeyTemplate != "":
		generator = keygen.NewMetadata(p.logger, cfg.CacheKeyTemplate, p.Metadata)
		if err := generator.Check(); err != nil {
			return fmt.Errorf("parse failed, falling back to default, %w", err)
		}

		options = append(options, cache.WithFallbackGenerator(keygen.NewHash(p.Metadata.Commit.Branch)))
	default:
		{
			generator = keygen.NewHash(p.Metadata.Commit.Branch)
			options = append(options, cache.WithFallbackGenerator(keygen.NewStatic(p.Metadata.Commit.Branch)))
		}
	}

	options = append(options, cache.WithOverride(p.Config.Override),
		cache.WithFailRestoreIfKeyNotPresent(p.Config.FailRestoreIfKeyNotPresent), cache.WithEnableCacheKeySeparator(p.Config.EnableCacheKeySeparator))

	// 2. Initialize storage backend.
	b, err := backend.FromConfig(p.logger, cfg.Backend, backend.Config{
		Debug:      cfg.Debug,
		Azure:      cfg.Azure,
		FileSystem: cfg.FileSystem,
		GCS:        cfg.GCS,
		S3:         cfg.S3,
		SFTP:       cfg.SFTP,
		Harness:    cfg.Harness,
	})
	if err != nil {
		return fmt.Errorf("initialize backend <%s>, %w", cfg.Backend, err)
	}

	// 3. Initialize cache.
	c := cache.New(p.logger,
		storage.New(p.logger, b, cfg.StorageOperationTimeout),
		archive.FromFormat(p.logger, localRoot, cfg.ArchiveFormat,
			archive.WithSkipSymlinks(cfg.SkipSymlinks),
			archive.WithCompressionLevel(cfg.CompressionLevel),
		),
		generator,
		cfg.Backend,
		cfg.AccountID,
		options...,
	)

	// 4. Expand the mount paths.
	p.Config.Mount = expandConfigPath(p.Config.Mount)

	// 5. Select mode
	if cfg.Rebuild {
		if err := c.Rebuild(p.Config.Mount); err != nil {
			level.Debug(p.logger).Log("err", fmt.Sprintf("%+v\n", err))
			return Error(fmt.Sprintf("[IMPORTANT] build cache, %+v\n", err))
		}
	}

	if cfg.Restore {
		if err := c.Restore(p.Config.Mount); err != nil {
			level.Debug(p.logger).Log("err", fmt.Sprintf("%+v\n", err))
			return Error(fmt.Sprintf("[IMPORTANT] restore cache, %+v\n", err))
		}
	}

	// FLUSH

	return nil
}
