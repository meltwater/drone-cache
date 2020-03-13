package backend

import (
	"io"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/meltwater/drone-cache/cache"
	"github.com/pkg/errors"
)

var _ (cache.Backend) = (*alibabaOSSBackend)(nil)

// alibabaOssBackend is the alibaba Oss Cloud implementation of the Backend
type alibabaOSSBackend struct {
	bucket     string
	acl        string
	encryption string
	client     *oss.Client
}

func newAlibabaOss(bucket string, conf *oss.Config, opts ...oss.ClientOption) (cache.Backend, error) {
	client, err := oss.New(conf.Endpoint, conf.AccessKeyID, conf.AccessKeySecret, opts...)
	if err != nil {
		return nil, errors.Wrap(err, "could not create alibabaOSS client")
	}
	return &alibabaOSSBackend{
		bucket: bucket,
		client: client,
	}, nil
}

func (c alibabaOSSBackend) Get(p string) (io.ReadCloser, error) {
	bucket, err := c.client.Bucket(c.bucket)
	if err != nil {
		return nil, errors.Wrap(err, "couldn't get the object")
	}
	reader, err := bucket.GetObject(p)
	if err != nil {
		return nil, errors.Wrap(err, "couldn't get the object")
	}
	return reader, nil
}

func (c alibabaOSSBackend) Put(p string, src io.ReadSeeker) error {
	bucket, err := c.client.Bucket(c.bucket)
	if err != nil {
		return errors.Wrap(err, "couldn't put the object")
	}

	options := []oss.Option{}

	if c.encryption != "" {
		option := oss.ServerSideEncryption(c.encryption)
		options = append(options, option)
	}

	if c.acl != "" {
		option := oss.ObjectACL(oss.ACLType(c.acl))
		options = append(options, option)

	}
	if err := bucket.PutObject(p, src, options...); err != nil {
		return errors.Wrap(err, "couldn't put the object")
	}
	return nil
}
