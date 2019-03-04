package backend

import (
	"fmt"
	"log"
	"os"
	"path"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/meltwater/drone-cache/cache"
	"github.com/pkg/errors"
)

type S3Config struct {
	// Indicates the files ACL, which should be one
	// of the following:
	//     private
	//     public-read
	//     public-read-write
	//     authenticated-read
	//     bucket-owner-read
	//     bucket-owner-full-control
	ACL        string
	Bucket     string
	Encryption string // if not "", enables server-side encryption. valid values are: AES256, aws:kms
	Endpoint   string
	Key        string

	// us-east-1
	// us-west-1
	// us-west-2
	// eu-west-1
	// ap-southeast-1
	// ap-southeast-2
	// ap-northeast-1
	// sa-east-1
	Region string
	Secret string

	PathStyle bool // Use path style instead of domain style. Should be true for minio and false for AWS
}

type FileSystemConfig struct {
	CacheRoot string
}

func InitializeS3Backend(c S3Config, debug bool) (cache.Backend, error) {
	awsConf := &aws.Config{
		Region:           aws.String(c.Region),
		Endpoint:         &c.Endpoint,
		DisableSSL:       aws.Bool(!strings.HasPrefix(c.Endpoint, "https://")),
		S3ForcePathStyle: aws.Bool(c.PathStyle),
	}

	if c.Key != "" && c.Secret != "" {
		awsConf.Credentials = credentials.NewStaticCredentials(c.Key, c.Secret, "")
	} else {
		log.Println("aws key and/or Secret not provided (falling back to anonymous credentials)")
	}

	if debug {
		log.Printf("[DEBUG] s3 backend config: %+v", c)
		awsConf.WithLogLevel(aws.LogDebugWithHTTPBody)
	}

	return newS3(c.Bucket, c.ACL, c.Encryption, awsConf), nil
}

func InitializeFileSystemBackend(c FileSystemConfig, debug bool) (cache.Backend, error) {
	if strings.TrimRight(path.Clean(c.CacheRoot), "/") == "" {
		return nil, fmt.Errorf("could not use <%s> as cache root, empty or root path given", c.CacheRoot)
	}

	if _, err := os.Stat(c.CacheRoot); err != nil {
		msg := fmt.Sprintf("could not use <%s> as cache root, make sure volume is mounted", c.CacheRoot)
		return nil, errors.Wrap(err, msg)
	}

	if debug {
		log.Printf("[DEBUG] filesystem backend config: %+v", c)
	}

	return newFileSystem(c.CacheRoot), nil
}
