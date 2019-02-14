// Package plugin for caching directories using given backends
package plugin

import (
	"crypto/md5"
	"fmt"
	"io"
	"log"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/pkg/errors"

	"github.com/meltwater/drone-s3-cache/cache"
	"github.com/meltwater/drone-s3-cache/cache/backend"
)

type (
	// Repo stores information about repository that is built
	Repo struct {
		Owner   string
		Name    string
		Link    string
		Avatar  string
		Branch  string
		Private bool
		Trusted bool
	}

	// Build stores information about current build
	Build struct {
		Number   int
		Event    string
		Status   string
		Deploy   string
		Created  int64
		Started  int64
		Finished int64
		Link     string
	}

	// Commit stores information about current commit
	Commit struct {
		Remote  string
		Sha     string
		Ref     string
		Link    string
		Branch  string
		Message string
		Author  Author
	}

	// Author stores information about current commit's author
	Author struct {
		Name   string
		Email  string
		Avatar string
	}

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
		ACL     string
		Branch  string
		Bucket  string
		Default string // default master branch
		// if not "", enable server-side encryption
		// valid values are:
		//     AES256
		//     aws:kms
		Encryption string
		Endpoint   string
		Key        string
		Mount      []string
		// Use path style instead of domain style
		// Should be true for minio and false for AWS
		PathStyle bool
		Rebuild   bool
		Region    string
		Repo      string
		Restore   bool
		Secret    string
	}

	// Plugin stores metadata about current plugin
	Plugin struct {
		Repo   Repo
		Build  Build
		Commit Commit
		Config Config
	}
)

// Exec entry point of Plugin, where the magic happens
func (p *Plugin) Exec() error {
	c := p.Config
	if c.Rebuild && c.Restore {
		return errors.New("rebuild and restore are mutually exclusive, please set only one of them")
	}

	conf := &aws.Config{
		Region:           aws.String(c.Region),
		Endpoint:         &c.Endpoint,
		DisableSSL:       aws.Bool(!strings.HasPrefix(c.Endpoint, "https://")),
		S3ForcePathStyle: aws.Bool(c.PathStyle),
	}
	// allowing to use the instance role or provide a key and secret
	if c.Key != "" && c.Secret != "" {
		conf.Credentials = credentials.NewStaticCredentials(c.Key, c.Secret, "")
	}

	cacheBackend := backend.NewS3(c.Bucket, c.ACL, c.Encryption, conf)

	if c.Rebuild {
		if err := p.processRebuild(cacheBackend); err != nil {
			return errors.Wrap(err, "process rebuild failed")
		}
	}

	if c.Restore {
		if err := p.processRestore(cacheBackend); err != nil {
			return errors.Wrap(err, "process restore failed")
		}
	}

	return nil
}

// Helpers

// processRebuild the remote cache from the local environment
func (p Plugin) processRebuild(b cache.Backend) error {
	c := p.Config
	now := time.Now()

	for _, mount := range c.Mount {
		cacheKey := hash(mount, c.Branch)
		path := filepath.Join(c.Repo, cacheKey)

		log.Printf("rebuilding cache for directory <%s> to remote cache <%s>", mount, path)
		if err := cache.Upload(b, mount, path); err != nil {
			return errors.Wrap(err, "could not upload")
		}
	}
	log.Printf("cache built in %v", time.Since(now))
	return nil
}

// processRestore the local environment from the remote cache
func (p Plugin) processRestore(b cache.Backend) error {
	c := p.Config
	now := time.Now()

	for _, mount := range c.Mount {
		cacheKey := hash(mount, c.Branch)
		path := filepath.Join(c.Repo, cacheKey)

		log.Printf("restoring directory <%s> from remote cache <%s>", mount, path)
		if err := cache.Download(b, path, mount); err != nil {
			return errors.Wrap(err, "could not download")
		}
	}
	log.Printf("cache restored in %v", time.Since(now))
	return nil
}

// hash a file name based on path and branch
func hash(mount, branch string) string {
	parts := []string{mount, branch}

	// calculate the hash using the branch
	h := md5.New()
	for _, part := range parts {
		io.WriteString(h, part)
	}
	return fmt.Sprintf("%x", h.Sum(nil))
}
