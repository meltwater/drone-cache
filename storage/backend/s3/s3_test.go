// +build integration

package s3

import (
	"bytes"
	"context"
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/go-kit/kit/log"

	"github.com/meltwater/drone-cache/test"
)

const (
	defaultEndpoint        = "127.0.0.1:9000"
	defaultAccessKey       = "AKIAIOSFODNN7EXAMPLE"
	defaultSecretAccessKey = "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY"
	defaultRegion          = "eu-west-1"
	defaultACL             = "private"
)

var (
	endpoint        = getEnv("TEST_S3_ENDPOINT", defaultEndpoint)
	accessKey       = getEnv("TEST_S3_ACCESS_KEY", defaultAccessKey)
	secretAccessKey = getEnv("TEST_S3_SECRET_KEY", defaultSecretAccessKey)
	acl             = getEnv("TEST_S3_ACL", defaultACL)
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
	client := newClient()
	bucket := "s3-round-trip"

	_, err := client.CreateBucketWithContext(context.Background(), &s3.CreateBucketInput{
		Bucket: aws.String(bucket),
	})
	test.Ok(t, err)

	b, err := New(
		log.NewNopLogger(),
		Config{
			ACL:       acl,
			Bucket:    bucket,
			Endpoint:  endpoint,
			Key:       accessKey,
			PathStyle: true, // Should be true for minio and false for AWS.
			Region:    defaultRegion,
			Secret:    secretAccessKey,
		},
		false,
	)
	test.Ok(t, err)

	return b, func() {
		_, _ = client.DeleteBucket(&s3.DeleteBucketInput{
			Bucket: aws.String(bucket),
		})
	}
}

func newClient() *s3.S3 {
	conf := &aws.Config{
		Region:           aws.String(defaultRegion),
		Endpoint:         aws.String(endpoint),
		DisableSSL:       aws.Bool(!strings.HasPrefix(endpoint, "https://")),
		S3ForcePathStyle: aws.Bool(true),
		Credentials:      credentials.NewStaticCredentials(accessKey, secretAccessKey, ""),
	}

	return s3.New(session.Must(session.NewSessionWithOptions(session.Options{})), conf)
}

func getEnv(key, defaultVal string) string {
	value, ok := os.LookupEnv(key)
	if !ok {
		return defaultVal
	}

	return value
}
