package backend

import (
	"context"
	"fmt"
	"io"
	"net/url"

	"github.com/Azure/azure-storage-blob-go/azblob"
	"github.com/meltwater/drone-cache/cache"
)

type azureBackend struct {
	containerURL azblob.ContainerURL
	ctx          context.Context
}

func newAzure(Credential *azblob.SharedKeyCredential, url *url.URL, c *AzureConfig) cache.Backend {

	pipeline := azblob.NewPipeline(Credential, azblob.PipelineOptions{})

	containerURL := azblob.NewContainerURL(*url, pipeline)
	ctx := context.Background()

	// Always creating new container, it will throw error if it already exists
	containerURL.Create(ctx, azblob.Metadata{}, azblob.PublicAccessNone)

	return &azureBackend{
		containerURL: containerURL,
		ctx:          ctx,
	}
}

func (c *azureBackend) Get(p string) (io.ReadCloser, error) {

	blobURL := c.containerURL.NewBlockBlobURL(p)

	downloadResponse, err := blobURL.Download(c.ctx, 0, azblob.CountToEnd, azblob.BlobAccessConditions{}, false)
	if err != nil {
		return nil, fmt.Errorf("get the object %w", err)
	}

	// NOTE: automatically retries are performed if the connection fails
	bodyStream := downloadResponse.Body(azblob.RetryReaderOptions{MaxRetryRequests: 20})

	return bodyStream, nil
}

// Put uploads the contents of the io.ReadSeeker
func (c *azureBackend) Put(p string, src io.ReadSeeker) error {

	blobURL := c.containerURL.NewBlockBlobURL(p)

	fmt.Printf("Uploading the file with blob name: %s\n", p)
	_, err := blobURL.Upload(c.ctx, src, azblob.BlobHTTPHeaders{}, azblob.Metadata{}, azblob.BlobAccessConditions{})
	if err != nil {
		return fmt.Errorf("put the object %w", err)
	}

	return nil
}
