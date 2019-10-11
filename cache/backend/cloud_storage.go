package backend

import (
	"context"
	"io"

	"github.com/meltwater/drone-cache/cache"

	"cloud.google.com/go/storage"
	"google.golang.org/api/option"
)

// cloudStorageBackend is an Cloud Storage implementation of the Backend
type cloudStorageBackend struct {
	bucket     string
	acl        string
	encryption string
	client     *storage.Client
}

// newCloudStorage returns a new Cloud Storage remote Backend implemented
func newCloudStorage(bucket, acl, encryption string, opts ...option.ClientOption) (cache.Backend, error) {
	ctx := context.Background()
	client, err := storage.NewClient(ctx, opts...)
	if err != nil {
		return nil, err
	}
	return &cloudStorageBackend{
		bucket:     bucket,
		acl:        acl,
		encryption: encryption,
		client:     client,
	}, nil
}

// Get returns an io.Reader for reading the contents of the file
func (c *cloudStorageBackend) Get(p string) (io.ReadCloser, error) {
	bkt := c.client.Bucket(c.bucket)
	obj := bkt.Object(p)
	if c.encryption != "" {
		obj = obj.Key([]byte(c.encryption))
	}

	// TODO: use a timeout or cancel context
	return obj.NewReader(context.TODO())
}

// Put uploads the contents of the io.ReadSeeker
func (c *cloudStorageBackend) Put(p string, src io.ReadSeeker) error {
	bkt := c.client.Bucket(c.bucket)
	obj := bkt.Object(p)
	if c.encryption != "" {
		obj = obj.Key([]byte(c.encryption))
	}

	//TODO: use a timeout or cancel context
	w := obj.NewWriter(context.TODO())
	_, err := io.Copy(w, src)
	return err
}
