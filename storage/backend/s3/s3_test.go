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
	"github.com/go-kit/log"

	"github.com/meltwater/drone-cache/test"
)

const (
	defaultEndpoint            = "127.0.0.1:9000"
	defaultAccessKey           = "AKIAIOSFODNN7EXAMPLE"
	defaultSecretAccessKey     = "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY"
	defaultRegion              = "eu-west-1"
	defaultACL             	   = "private"
	defaultUserAccessKey   	   = "foo"
	defaultUserSecretAccessKey = "barbarbar"
)

var (
	endpoint        	= getEnv("TEST_S3_ENDPOINT", defaultEndpoint)
	accessKey       	= getEnv("TEST_S3_ACCESS_KEY", defaultAccessKey)
	secretAccessKey 	= getEnv("TEST_S3_SECRET_KEY", defaultSecretAccessKey)
	acl             	= getEnv("TEST_S3_ACL", defaultACL)
	userAccessKey       = getEnv("TEST_USER_S3_ACCESS_KEY", defaultUserAccessKey)
	userSecretAccessKey = getEnv("TEST_USER_S3_SECRET_KEY", defaultUserSecretAccessKey)
)

func TestRoundTrip(t *testing.T) {
	t.Parallel()

	backend, cleanUp := setup(t, Config{
		ACL:       acl,
		Bucket:    "s3-round-trip",
		Endpoint:  endpoint,
		Key:       accessKey,
		PathStyle: true, // Should be true for minio and false for AWS.
		Region:    defaultRegion,
		Secret:    secretAccessKey,
		DisableSSL:true // minio unable to handle https requests
	})
	t.Cleanup(cleanUp)
	roundTrip(t, backend)
}

func TestRoundTripWithAssumeRole(t *testing.T) {
	t.Parallel()

	backend, cleanUp := setup(t, Config{
		ACL:       acl,
		Bucket:    "s3-round-trip-with-role",
		Endpoint:  endpoint,
		StsEndpoint: endpoint,
		Key:       userAccessKey,
		PathStyle: true, // Should be true for minio and false for AWS.
		Region:    defaultRegion,
		Secret:    userSecretAccessKey,
		RoleArn:   "arn:aws:iam::account-id:role/TestRole",
		DisableSSL:true // minio unable to handle https requests
	})
	t.Cleanup(cleanUp)
	roundTrip(t, backend)
}

func roundTrip(t *testing.T, backend *Backend) {
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

func setup(t *testing.T, config Config) (*Backend, func()) {
	client := newClient(config)

	_, err := client.CreateBucketWithContext(context.Background(), &s3.CreateBucketInput{
		Bucket: aws.String(config.Bucket),
	})
	test.Ok(t, err)

	b, err := New(
		log.NewNopLogger(),
		config,
		false,
	)
	test.Ok(t, err)

	return b, func() {
		_, err = client.DeleteBucket(&s3.DeleteBucketInput{
			Bucket: aws.String(config.Bucket),
		})
	}
}

func newClient(config Config) *s3.S3 {
	conf := &aws.Config{
		Region:           aws.String(defaultRegion),
		Endpoint:         aws.String(endpoint),
		DisableSSL:       aws.Bool(!strings.HasPrefix(endpoint, "https://")),
		S3ForcePathStyle: aws.Bool(true),
		Credentials:      credentials.NewStaticCredentials(config.Key, config.Secret, ""),
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
