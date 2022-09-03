package alioss

import (
	"context"
	"fmt"
	"io"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/pkg/errors"
)

// Backend implements storage.Backend for AWs S3.
type Backend struct {
	logger log.Logger

	bucket     string
	acl        string
	encryption string
	client     *oss.Client
}

func newAlibabaOss(bucket string, conf *oss.Config, opts ...oss.ClientOption) (*Backend, error) {
	client, err := oss.New(conf.Endpoint, conf.AccessKeyID, conf.AccessKeySecret, opts...)
	if err != nil {
		return nil, errors.Wrap(err, "could not create alibabaOSS client")
	}

	return &Backend{
		bucket: bucket,
		client: client,
	}, nil
}

func New(l log.Logger, c Config, debug bool) (*Backend, error) {
	ossConf := &oss.Config{}
	if c.Endpoint != "" {
		ossConf.Endpoint = c.Endpoint
	}

	if c.AccesKeyID != "" {
		ossConf.AccessKeyID = c.AccesKeyID
	}

	if c.AccesKeySecret != "" {
		ossConf.AccessKeySecret = c.AccesKeySecret
	}

	if debug {
		level.Debug(l).Log("msg", "oss storage backend", "config", fmt.Sprintf("%+v", c))
	}

	return newAlibabaOss(c.Bucket, ossConf)
}

func (c Backend) Get(ctx context.Context, p string, w io.Writer) error {
	bucket, err := c.client.Bucket(c.bucket)
	if err != nil {
		return errors.Wrap(err, "couldn't get the object")
	}

	reader, err := bucket.GetObject(p)
	if err != nil {
		return errors.Wrap(err, "couldn't get the object")
	}

	return nil
}

func (c Backend) Put(ctx context.Context, p string, src io.Reader) error {
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

func (c Backend) Exists(ctx context.Context, p string) (bool, error) {
	bucket, err := c.client.Bucket(c.bucket)
	if err != nil {
		return false, errors.Wrap(err, "couldn't get the bucket object")
	}

	options := []oss.Option{}

	result, err := bucket.IsObjectExist(p, options...)
	if err != nil {
		return false, errors.Wrap(err, "couldn't get the object")
	}

	return result, nil
}
