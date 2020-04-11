package s3

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/meltwater/drone-cache/internal"
)

// Backend TODO
type Backend struct {
	logger log.Logger

	bucket     string
	acl        string
	encryption string
	client     *s3.S3
}

// New creates an S3 backend.
func New(l log.Logger, c Config, debug bool) (*Backend, error) {
	conf := &aws.Config{
		Region:           aws.String(c.Region),
		Endpoint:         &c.Endpoint,
		DisableSSL:       aws.Bool(!strings.HasPrefix(c.Endpoint, "https://")),
		S3ForcePathStyle: aws.Bool(c.PathStyle),
	}

	if c.Key != "" && c.Secret != "" {
		conf.Credentials = credentials.NewStaticCredentials(c.Key, c.Secret, "")
	} else {
		level.Warn(l).Log("msg", "aws key and/or Secret not provided (falling back to anonymous credentials)")
	}

	level.Debug(l).Log("msg", "s3 backend", "config", fmt.Sprintf("%#v", c))

	if debug {
		conf.WithLogLevel(aws.LogDebugWithHTTPBody)
	}

	client := s3.New(session.Must(session.NewSessionWithOptions(session.Options{})), conf)

	return &Backend{
		logger:     l,
		bucket:     c.Bucket,
		acl:        c.ACL,
		encryption: c.Encryption,
		client:     client,
	}, nil
}

// Get writes downloaded content to the given writer.
func (b *Backend) Get(ctx context.Context, p string, w io.Writer) error {
	in := &s3.GetObjectInput{
		Bucket: aws.String(b.bucket),
		Key:    aws.String(p),
	}

	errCh := make(chan error)

	go func() {
		defer close(errCh)

		out, err := b.client.GetObjectWithContext(ctx, in)
		if err != nil {
			errCh <- fmt.Errorf("get the object, %w", err)
			return
		}

		defer internal.CloseWithErrLogf(b.logger, out.Body, "response body, close defer")

		_, err = io.Copy(w, out.Body)
		if err != nil {
			errCh <- fmt.Errorf("copy the object, %w", err)
		}
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}

// Put uploads contents of the given reader.
func (b *Backend) Put(ctx context.Context, p string, r io.Reader) error {
	var (
		uploader = s3manager.NewUploaderWithClient(b.client)
		in       = &s3manager.UploadInput{
			Bucket: aws.String(b.bucket),
			Key:    aws.String(p),
			ACL:    aws.String(b.acl),
			Body:   r,
		}
	)

	if b.encryption != "" {
		in.ServerSideEncryption = aws.String(b.encryption)
	}

	if _, err := uploader.UploadWithContext(ctx, in); err != nil {
		return fmt.Errorf("put the object, %w", err)
	}

	return nil
}

// Exists checks if object already exists.
func (b *Backend) Exists(ctx context.Context, p string) (bool, error) {
	in := &s3.HeadObjectInput{
		Bucket: aws.String(b.bucket),
		Key:    aws.String(p),
	}

	out, err := b.client.HeadObjectWithContext(ctx, in)
	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok && awsErr.Code() == s3.ErrCodeNoSuchKey || awsErr.Code() == "NotFound" {
			return false, nil
		}

		return false, fmt.Errorf("head the object, %w", err)
	}

	// Normally if file not exists it will be already detected by error above but in some cases
	// Minio can return success status for without ETag, detect that here.
	return *out.ETag != "", nil
}
