package provider

import (
	"io"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/meltwater/drone-s3-cache/cache"
)

// s3provider is an S3 implementation of the Provider.
type s3provider struct {
	bucket     string
	acl        string
	encryption string
	client     *s3.S3
}

// NewS3 returns a new SFTP remote Provider implemented.
func NewS3(bucket, acl, encryption string, conf *aws.Config) cache.Provider {
	client := s3.New(session.New(), conf)
	return &s3provider{
		bucket:     bucket,
		acl:        acl,
		encryption: encryption,
		client:     client,
	}
}

// Get returns an io.Reader for reading the contents of the file.
func (c *s3provider) Get(p string) (io.ReadCloser, error) {
	out, err := c.client.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(c.bucket),
		Key:    aws.String(p),
	})
	if err != nil {
		return nil, err
	}
	return out.Body, nil
}

// Put uploads the contents of the io.ReadSeeker
func (c *s3provider) Put(p string, src io.ReadSeeker) error {
	in := &s3.PutObjectInput{
		Bucket: aws.String(c.bucket),
		Key:    aws.String(p),
		ACL:    aws.String(c.acl),
		Body:   src,
	}
	if c.encryption != "" {
		in.ServerSideEncryption = aws.String(c.encryption)
	}
	_, err := c.client.PutObject(in)
	return err
}
