// +build integration

package azure

import (
	"bytes"
	"context"
	"io/ioutil"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/meltwater/drone-cache/test"
)

const (
	defaultBlobStorageURL = "127.0.0.1:10000"
	defaultAccountName    = "devstoreaccount1"
	defaultAccountKey     = "Eby8vdM02xNOcqFlqUwJPLlmEtlCDXJ1OUzFT50uSRZ6IFsuFq2UVErCz4I6tq/K1SZFPTOtr/KBHBeksoGMGw=="
	defaultContainerName  = "testcontainer"
)

var (
	blobURL       = getEnv("TEST_AZURITE_URL", defaultBlobStorageURL)
	accountName   = getEnv("TEST_ACCOUNT_NAME", defaultAccountName)
	accountKey    = getEnv("TEST_ACCOUNT_KEY", defaultAccountKey)
	containerName = getEnv("TEST_CONTAINER_NAME", defaultContainerName)
)

func TestRoundTrip(t *testing.T) {
	t.Parallel()

	backend, cleanUp := setup(t)
	t.Cleanup(cleanUp)

	content := "Hello world4"

	// Test Put
	test.Ok(t, backend.Put(context.TODO(), "test.t", strings.NewReader(content)))

	// Test Get
	var buf bytes.Buffer
	test.Ok(t, backend.Get(context.TODO(), "test.t", &buf))

	b, err := ioutil.ReadAll(&buf)
	test.Ok(t, err)

	test.Equals(t, []byte(content), b)

	exists, err := backend.Exists(context.TODO(), "test.t")
	test.Ok(t, err)

	test.Equals(t, true, exists)
}

// Helpers

func setup(t *testing.T) (*Backend, func()) {
	b, err := New(
		log.NewNopLogger(),
		Config{
			AccountName:    accountName,
			AccountKey:     accountKey,
			ContainerName:  containerName,
			BlobStorageURL: blobURL,
			Azurite:        true,
			Timeout:        30 * time.Second,
		},
	)
	test.Ok(t, err)

	return b, func() {}
}

func getEnv(key, defaultVal string) string {
	value, ok := os.LookupEnv(key)
	if !ok {
		return defaultVal
	}
	return value
}
