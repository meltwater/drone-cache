// Package plugin for caching directories using given backends
package plugin

import (
	"crypto/md5"
	"fmt"
	"io"
	"log"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/pkg/errors"

	"github.com/meltwater/drone-cache/cache"
	"github.com/meltwater/drone-cache/cache/backend"
)

type (
	// Repo stores information about repository that is built
	Repo struct {
		Avatar  string
		Branch  string
		Link    string
		Name    string
		Owner   string
		Private bool
		Trusted bool
	}

	// Build stores information about current build
	Build struct {
		Created  int64
		Deploy   string
		Event    string
		Finished int64
		Link     string
		Number   int
		Started  int64
		Status   string
	}

	// Commit stores information about current commit
	Commit struct {
		Author  Author
		Branch  string
		Link    string
		Message string
		Ref     string
		Remote  string
		Sha     string
	}

	// Author stores information about current commit's author
	Author struct {
		Avatar string
		Email  string
		Name   string
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
		ACL           string
		ArchiveFormat string
		Bucket        string
		CacheKey      string
		Debug         bool
		// if not "", enable server-side encryption
		// valid values are:
		//     AES256
		//     aws:kms
		Encryption string
		Endpoint   string
		Key        string
		Mount      []string
		// Use path style instead of domain style.
		// Should be true for minio and false for AWS
		PathStyle bool
		Rebuild   bool
		// us-east-1
		// us-west-1
		// us-west-2
		// eu-west-1
		// ap-southeast-1
		// ap-southeast-2
		// ap-northeast-1
		// sa-east-1
		Region  string
		Restore bool
		Secret  string
	}

	// Plugin stores metadata about current plugin
	Plugin struct {
		Build  Build
		Commit Commit
		Config Config
		Repo   Repo
	}
)

// Exec entry point of Plugin, where the magic happens
func (p *Plugin) Exec() error {
	c := p.Config

	// 1. Check paramaters
	if c.Debug {
		log.Println("DEBUG MODE enabled!")
	}

	if c.Rebuild && c.Restore {
		return errors.New("rebuild and restore are mutually exclusive, please set only one of them")
	}

	_, err := template.New("cacheKey").Parse(p.Config.CacheKey)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("could not parse <%s> as cache key template, falling back to default", p.Config.CacheKey))
	}

	// 2. Initialize backend
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
	backend := backend.NewS3(c.Bucket, c.ACL, c.Encryption, awsConf)

	// 3. Initialize cache
	cch := cache.New(backend, c.ArchiveFormat)

	// 4. Select mode
	if c.Rebuild {
		if err := p.processRebuild(cch); err != nil {
			log.Printf("WARNING: could not build cache. process rebuild failed, %+v\n", err)
			return nil
		}
	}

	if c.Restore {
		if err := p.processRestore(cch); err != nil {
			log.Printf("WARNING: could not restore cache. process restore failed, %+v\n", err)
			return nil
		}
	}

	return nil
}

// processRebuild the remote cache from the local environment
func (p Plugin) processRebuild(c cache.Cache) error {
	now := time.Now()
	branch := p.Commit.Branch

	for _, mount := range p.Config.Mount {
		key, err := p.cacheKey(mount)
		if err != nil {
			log.Printf("%v, falling back to default key\n", err)
			key = hash(mount, branch)
		}
		path := filepath.Join(p.Repo.Name, key)

		log.Printf("rebuilding cache for directory <%s> to remote cache <%s>\n", mount, path)
		if err := c.Upload(mount, path); err != nil {
			return errors.Wrap(err, "could not upload")
		}
	}
	log.Printf("cache built in %v", time.Since(now))
	return nil
}

// processRestore the local environment from the remote cache
func (p Plugin) processRestore(c cache.Cache) error {
	now := time.Now()
	branch := p.Commit.Branch

	for _, mount := range p.Config.Mount {
		key, err := p.cacheKey(mount)
		if err != nil {
			log.Printf("%v, falling back to default key\n", err)
			key = hash(mount, branch)
		}
		path := filepath.Join(p.Repo.Name, key)

		log.Printf("restoring directory <%s> from remote cache <%s>\n", mount, path)
		if err := c.Download(path, mount); err != nil {
			return errors.Wrap(err, "could not download")
		}
	}
	log.Printf("cache restored in %v", time.Since(now))
	return nil
}

// cacheKey generates key from given template as parameter or fallbacks hash
func (p Plugin) cacheKey(mount string) (string, error) {
	if p.Config.CacheKey == "" {
		return "", errors.New("cache key template is empty")
	}

	log.Println("using provided cache key template")
	t, err := template.New("cacheKey").Parse(p.Config.CacheKey)
	if err != nil {
		return "", errors.Wrap(err, fmt.Sprintf("could not parse <%s> as cache key template, falling back to default\n", p.Config.CacheKey))
	}

	data := struct {
		Build  Build
		Commit Commit
		Repo   Repo
	}{
		Build:  p.Build,
		Commit: p.Commit,
		Repo:   p.Repo,
	}

	var b strings.Builder
	err = t.Execute(&b, data)
	if err != nil {
		return "", errors.Wrap(err, fmt.Sprintf("could not build <%s> as cache key, falling back to default. %+v\n", p.Config.CacheKey, err))
	}

	return filepath.Join(b.String(), mount), nil
}

// Helpers

// hash generates a key based on filename paths and branch
func hash(mount, branch string) string {
	parts := []string{mount, branch}

	// calculate the hash using the branch
	h := md5.New()
	for _, part := range parts {
		io.WriteString(h, part)
	}
	return fmt.Sprintf("%x", h.Sum(nil))
}
