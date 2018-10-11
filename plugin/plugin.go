package plugin

import (
	"crypto/md5"
	"fmt"
	"io"
	"log"
	"path/filepath"
	"strings"
	"time"

	"github.com/meltwater/drone-s3-cache/provider"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/meltwater/drone-s3-cache/cache"
)

// Plugin for caching directories using given Providers.
type Plugin struct {
	Rebuild bool
	Restore bool
	Mount   []string

	Endpoint string
	Key      string
	Secret   string
	Bucket   string
	Region   string

	// if not "", enable server-side encryption
	// valid values are:
	//     AES256
	//     aws:kms
	Encryption string

	// Indicates the files ACL, which should be one
	// of the following:
	//     private
	//     public-read
	//     public-read-write
	//     authenticated-read
	//     bucket-owner-read
	//     bucket-owner-full-control
	ACL string

	// Use path style instead of domain style.
	//
	// Should be true for minio and false for AWS.
	PathStyle bool

	Repo    string
	Branch  string
	Default string // default master branch
}

// Exec entry point of Plugin, where the magic happens.
func (p *Plugin) Exec() error {
	conf := &aws.Config{
		Region:           aws.String(p.Region),
		Endpoint:         &p.Endpoint,
		DisableSSL:       aws.Bool(strings.HasPrefix(p.Endpoint, "http://")),
		S3ForcePathStyle: aws.Bool(p.PathStyle),
	}

	// Allowing to use the instance role or provide a key and secret
	if p.Key != "" && p.Secret != "" {
		conf.Credentials = credentials.NewStaticCredentials(p.Key, p.Secret, "")
	}

	cacheProvider := provider.NewS3(p.Bucket, p.ACL, p.Encryption, conf)
	if p.Rebuild {
		now := time.Now()
		if err := p.processRebuild(cacheProvider); err != nil {
			log.Println(err)
		} else {
			log.Printf("cache built in %v", time.Since(now))
		}
	}

	if p.Restore {
		now := time.Now()
		if err := p.processRestore(cacheProvider); err != nil {
			log.Println(err)
		} else {
			log.Printf("cache restored in %v", time.Since(now))
		}
	}

	return nil
}

// Helpers

// processRebuild the remote cache from the local environment.
func (p Plugin) processRebuild(c cache.Provider) error {
	for _, mount := range p.Mount {
		cacheKey := hash(mount, p.Branch)
		path := filepath.Join(p.Repo, cacheKey)

		log.Printf("archiving directory <%s> to remote cache <%s>", mount, path)
		err := cache.Upload(c, mount, path)
		if err != nil {
			return err
		}
	}
	return nil
}

// processRestore the local environment from the remote cache.
func (p Plugin) processRestore(c cache.Provider) error {
	for _, mount := range p.Mount {
		cacheKey := hash(mount, p.Branch)
		path := filepath.Join(p.Repo, cacheKey)

		log.Printf("restoring directory <%s> from remote cache <%s>", mount, path)
		err := cache.Download(c, path, mount)
		if err != nil {
			return err
		}
	}
	return nil
}

// hash a file name based on path and branch.
func hash(mount, branch string) string {
	parts := []string{mount, branch}

	// calculate the hash using the branch
	h := md5.New()
	for _, part := range parts {
		io.WriteString(h, part)
	}
	return fmt.Sprintf("%x", h.Sum(nil))
}
