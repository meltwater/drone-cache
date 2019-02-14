package backend

import (
	"io"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/pkg/errors"

	"github.com/meltwater/drone-s3-cache/cache"
)

// s3Backend is an S3 implementation of the Backend
type s3Backend struct {
	bucket     string
	acl        string
	encryption string
	client     *s3.S3
}

// NewS3 returns a new S3 remote Backend implemented
func NewS3(bucket, acl, encryption string, conf *aws.Config) cache.Backend {
	client := s3.New(session.New(), conf)
	return &s3Backend{
		bucket:     bucket,
		acl:        acl,
		encryption: encryption,
		client:     client,
	}
}

// Get returns an io.Reader for reading the contents of the file
func (c *s3Backend) Get(p string) (io.ReadCloser, error) {
	out, err := c.client.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(c.bucket),
		Key:    aws.String(p),
	})
	if err != nil {
		return nil, errors.Wrap(err, "couldn't get the object")
	}

	return out.Body, nil
}

// Put uploads the contents of the io.ReadSeeker
func (c *s3Backend) Put(p string, src io.ReadSeeker) error {
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
	return errors.Wrap(err, "couldn't put the object")
}
