// +build integration

package gcs

import (
	"bytes"
	"context"
	"crypto/tls"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	gcstorage "cloud.google.com/go/storage"
	"github.com/go-kit/kit/log"
	"github.com/meltwater/drone-cache/test"
	"google.golang.org/api/option"
)

const (
	defaultEndpoint   = "http://127.0.0.1:4443/storage/v1/"
	defaultPublicHost = "localhost:4443"
	defaultApiKey     = ""
	bucketName        = "gcs-round-trip"
)

var (
	endpoint   = getEnv("TEST_GCS_ENDPOINT", defaultEndpoint)
	apiKey     = getEnv("TEST_GCS_API_KEY", defaultApiKey)
	publicHost = getEnv("TEST_STORAGE_EMULATOR_HOST", defaultPublicHost)
)

func TestRoundTrip(t *testing.T) {
	t.Parallel()

	backend, cleanUp := setup(t)
	t.Cleanup(cleanUp)

	content := "Hello world4"

	// Test Put
	test.Ok(t, backend.Put(context.TODO(), "test.txt", strings.NewReader(content)))

	// Test Get
	backend = getBackend(t, bucketName) // This weird env set and unset dance has to be made to mak it work with GCP client.
	var buf bytes.Buffer
	test.Ok(t, backend.Get(context.Background(), "test.txt", &buf))

	b, err := ioutil.ReadAll(&buf)
	test.Ok(t, err)

	test.Equals(t, []byte(content), b)
}

// Helpers

func setup(t *testing.T) (*Backend, func()) {
	client := newClient(t)
	bucket := client.Bucket(bucketName)

	test.Ok(t, bucket.Create(context.Background(), "drone-cache", &gcstorage.BucketAttrs{}))

	return putBackend(t, bucketName), func() {
		_ = os.Unsetenv("STORAGE_EMULATOR_HOST")
		_ = bucket.Delete(context.Background())
		_ = client.Close()
	}
}

func putBackend(t *testing.T, bucketName string) *Backend {
	b, err := New(
		log.NewLogfmtLogger(os.Stdout),
		Config{
			Bucket:   bucketName,
			Endpoint: endpoint,
			APIKey:   apiKey,
			Timeout:  30 * time.Second,
		},
	)
	test.Ok(t, err)

	return b
}

func getBackend(t *testing.T, bucketName string) *Backend {
	// This weird env set and unset dance has to be made to mak it work with GCP client.
	if _, ok := os.LookupEnv("STORAGE_EMULATOR_HOST"); !ok {
		test.Ok(t, os.Setenv("STORAGE_EMULATOR_HOST", publicHost))
	}

	b, err := New(
		log.NewLogfmtLogger(os.Stdout),
		Config{
			Bucket:   bucketName,
			Endpoint: endpoint,
			APIKey:   apiKey,
			Timeout:  30 * time.Second,
		},
	)
	test.Ok(t, err)

	_ = os.Unsetenv("STORAGE_EMULATOR_HOST")
	return b
}

func newClient(t *testing.T) *gcstorage.Client {
	var opts []option.ClientOption

	if apiKey != "" {
		opts = append(opts, option.WithAPIKey(apiKey))
	} else {
		opts = append(opts, option.WithoutAuthentication())
	}

	if endpoint != "" {
		opts = append(opts, option.WithEndpoint(endpoint))
	}

	if !strings.HasPrefix(endpoint, "https://") {
		opts = append(opts, option.WithHTTPClient(&http.Client{Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, // ignore expired SSL certificates
		}}))
	}

	client, err := gcstorage.NewClient(context.Background(), opts...)
	test.Ok(t, err)

	return client
}

func getEnv(key, defaultVal string) string {
	value, ok := os.LookupEnv(key)
	if !ok {
		return defaultVal
	}

	return value
}
