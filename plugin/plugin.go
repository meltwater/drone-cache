// Package plugin for caching directories using given backends
package plugin

import (
	"fmt"
	"log"
	"path/filepath"
	"strings"
	"time"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/pkg/errors"

	"github.com/meltwater/drone-cache/cache"
	"github.com/meltwater/drone-cache/cache/backend"
	"github.com/meltwater/drone-cache/metadata"
	"github.com/meltwater/drone-cache/plugin/cachekey"
)

type (
	// Config plugin-specific parameters and secrets
	Config struct {
		// Indicates the files ACL, which should be one
		// of the following:
		//     private
		//     public-read
		//     public-read-write
		//     authenticated-read
		//     bucket-owner-read
		//     bucket-owner-full-control
		ACL           string
		ArchiveFormat string
		Bucket        string
		CacheKey      string
		CacheRoot     string
		Encryption    string // if not "", enables server-side encryption. valid values are: AES256, aws:kms
		Endpoint      string
		Key           string

		// us-east-1
		// us-west-1
		// us-west-2
		// eu-west-1
		// ap-southeast-1
		// ap-southeast-2
		// ap-northeast-1
		// sa-east-1
		Region string
		Secret string

		Debug     bool
		PathStyle bool // Use path style instead of domain style. Should be true for minio and false for AWS
		Rebuild   bool
		Restore   bool

		Mount []string
	}

	// Plugin stores metadata about current plugin
	Plugin struct {
		Metadata  metadata.Metadata
		Config Config
	}
)

// Exec entry point of Plugin, where the magic happens
func (p *Plugin) Exec() error {
	c := p.Config

	// 1. Check parameters
	if c.Debug {
		log.Println("DEBUG MODE enabled!")
	}

	if c.Rebuild && c.Restore {
		return errors.New("rebuild and restore are mutually exclusive, please set only one of them")
	}

	_, err := cachekey.ParseTemplate(c.CacheKey)
	if err != nil {
		return errors.Wrap(
			err,
			fmt.Sprintf("could not parse <%s> as cache key template, falling back to default", c.CacheKey),
		)
	}

	// 2. Initialize backend
	backend, err := initializeBackend(c)

	// 3. Initialize cache
	cch := cache.New(backend, c.ArchiveFormat)

	// 4. Select mode
	if c.Rebuild {
		if err := processRebuild(cch, p.Config.CacheKey, p.Config.Mount, p.Metadata); err != nil {
			log.Printf("WARNING: could not build cache. process rebuild failed, %v\n", err)
			return nil
		}
	}

	if c.Restore {
		if err := processRestore(cch, p.Config.CacheKey, p.Config.Mount, p.Metadata); err != nil {
			log.Printf("WARNING: could not restore cache. process restore failed, %v\n", err)
			return nil
		}
	}

	return nil
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

		log.Printf("rebuilding cache for directory <%s> to remote cache <%s>\n", mount, path)
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

		log.Printf("restoring directory <%s> from remote cache <%s>\n", mount, path)
		if err := c.Pull(path, mount); err != nil {
			return errors.Wrap(err, "could not download")
		}
	}
	log.Printf("cache restored in %v", time.Since(now))
	return nil
}

// cacheKey generates key from given template as parameter or fallbacks hash
func cacheKey(p metadata.Metadata, cacheKeyTmpl, mount, branch string) (string, error) {
	log.Println("using provided cache key template")
	key, err := cachekey.Generate(cacheKeyTmpl, mount, metadata.Metadata{
		Build:  p.Build,
		Commit: p.Commit,
		Repo:   p.Repo,
	})

	if err != nil {
		log.Printf("%v, falling back to default key\n", err)
		key, err = cachekey.Hash(mount, branch)
		if err != nil {
			return "", errors.Wrap(err, "could not generate hash key for mounted")
		}
	}

	return key, nil
}

// initializeBackend initializes backend using given configuration
func initializeBackend(c Config) (cache.Backend, error) {
	if c.CacheRoot != "" {
		log.Println("IMPORTANT: using filesystem as backend")

		// TODO: What happens if it's not mounted?
		if _, err := os.Stat(c.CacheRoot); err != nil {
			return nil, errors.Wrap(
				err,
				fmt.Sprintf("could not use <%s> as cache root, make sure volume is mounted", c.CacheRoot),
			)
		}

		return backend.NewFileSystem(c.CacheRoot), nil
	}

	log.Println("IMPORTANT: using aws s3 as backend")
	var cred *credentials.Credentials
	if c.Key != "" && c.Secret != "" {
		// allowing to use the instance role or provide a key and secret
		cred = credentials.NewStaticCredentials(c.Key, c.Secret, "")
	} else {
		cred = credentials.AnonymousCredentials
		log.Println("aws key and/or Secret not provided (falling back to anonymous credentials)")
	}
	awsConf := &aws.Config{
		Region:           aws.String(c.Region),
		Endpoint:         &c.Endpoint,
		DisableSSL:       aws.Bool(!strings.HasPrefix(c.Endpoint, "https://")),
		S3ForcePathStyle: aws.Bool(c.PathStyle),
		Credentials:      cred,
	}

	if c.Debug {
		awsConf.WithLogLevel(aws.LogDebugWithHTTPBody)
	}

	return backend.NewS3(c.Bucket, c.ACL, c.Encryption, awsConf), nil
}
