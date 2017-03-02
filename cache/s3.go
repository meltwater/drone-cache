package cache

import (
	"io"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

// cacher is an SFTP implementation of the Cache.
type cacher struct {
	bucket     string
	acl        string
	encryption string
	client     *s3.S3
}

// New returns a new SFTP remote Cache implementated.
func New(bucket, acl, encryption string, conf *aws.Config) Cache {
	client := s3.New(session.New(), conf)
	return &cacher{
		bucket:     bucket,
		acl:        acl,
		encryption: encryption,
		client:     client,
	}
}

// Get returns an io.Reader for reading the contents of the file.
func (c *cacher) Get(p string) (io.ReadCloser, error) {
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
func (c *cacher) Put(p string, src io.ReadSeeker) error {
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
