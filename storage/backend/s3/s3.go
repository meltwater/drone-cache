package s3

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/credentials/stscreds"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/meltwater/drone-cache/internal"
)

// Backend implements storage.Backend for AWs S3.
type Backend struct {
	logger log.Logger

	bucket     string
	acl        string
	encryption string
	client     *s3.S3
}

// New creates an S3 backend.
func New(l log.Logger, c Config, debug bool) (*Backend, error) {
	// Set SSL mode (enable/disable) using configuration flags
	sslMode := setSSLMode(l, c)

	conf := &aws.Config{
		Region:           aws.String(c.Region),
		Endpoint:         &c.Endpoint,
		DisableSSL:       aws.Bool(sslMode),
		S3ForcePathStyle: aws.Bool(c.PathStyle),
	}

	// Use anonymous credentials if the S3 bucket is public
	if c.Public {
		conf.Credentials = credentials.AnonymousCredentials
	}

	if c.Key != "" && c.Secret != "" {
		conf.Credentials = credentials.NewStaticCredentials(c.Key, c.Secret, "")
	} else {
		level.Warn(l).Log("msg", "aws key and/or Secret not provided (falling back to anonymous credentials)")
	}

	if c.RoleArn != "" {
		stsConf := conf
		if c.StsEndpoint != "" {
			stsConf = conf.Copy(&aws.Config{
				Endpoint:   &c.StsEndpoint,
				DisableSSL: aws.Bool(sslMode),
			})
		} else {
			stsConf.Endpoint = nil
			stsConf.DisableSSL = nil
		}

		conf.Credentials = credentials.NewStaticCredentials(c.Key, c.Secret, "")
		crds := assumeRole(l, stsConf, c.RoleArn)
		conf.Credentials = credentials.NewStaticCredentials(crds.AccessKeyID, crds.SecretAccessKey, crds.SessionToken)
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
		// nolint: wrapcheck
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
		// nolint: errorlint
		if awsErr, ok := err.(awserr.Error); ok && awsErr.Code() == s3.ErrCodeNoSuchKey || awsErr.Code() == "NotFound" {
			return false, nil
		}

		return false, fmt.Errorf("head the object, %w", err)
	}

	// Normally if file not exists it will be already detected by error above but in some cases
	// Minio can return success status for without ETag, detect that here.
	return *out.ETag != "", nil
}

func assumeRole(l log.Logger, c *aws.Config, roleArn string) credentials.Value {
	sess, err := session.NewSession(&aws.Config{
		Credentials:                   c.Credentials,
		Region:                        c.Region,
		Endpoint:                      c.Endpoint,
		DisableSSL:                    c.DisableSSL,
		CredentialsChainVerboseErrors: aws.Bool(true),
	})
	if err != nil {
		level.Error(l).Log("msg", "s3 backend", "assume-role", err.Error())
	}

	creds, err := stscreds.NewCredentials(sess, roleArn, func(p *stscreds.AssumeRoleProvider) {
		p.RoleSessionName = "drone-cache"
	}).Get()
	if err != nil {
		level.Error(l).Log("msg", "s3 backend", "assume-role", err.Error())
	}

	return creds
}

// Set the mode for SSL for S3 connectivity. Default mode is enabled.
// if EnableSSL flag was set to false, then return DisableSSL=true.
// if a custom stsEndpoint was specified without https, set mode to true (disableSSL=true)
//  enable SSL for all other conditions. set disableSSL=false

func setSSLMode(l log.Logger, c Config) bool {
	level.Info(l).Log("msg", "Setting SSL mode from config...")
	switch {
	case c.EnableSSL == false:
		return true
	case c.StsEndpoint != "" && !strings.HasPrefix(c.StsEndpoint, "https://"):
		return true
	default:
		return false
	}
}
