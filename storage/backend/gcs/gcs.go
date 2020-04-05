package gcs

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/meltwater/drone-cache/internal"

	gcstorage "cloud.google.com/go/storage"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
)

// Backend is an Cloud Storage implementation of the Backend.
type Backend struct {
	logger log.Logger

	bucket     string
	acl        string
	encryption string
	client     *gcstorage.Client
}

// New creates a Google Cloud Storage backend.
func New(l log.Logger, c Config) (*Backend, error) {
	var opts []option.ClientOption

	level.Debug(l).Log("msg", "gc storage backend", "config", fmt.Sprintf("%+v", c))

	if c.Endpoint != "" {
		opts = append(opts, option.WithEndpoint(c.Endpoint))
	}

	if !strings.HasPrefix(c.Endpoint, "https://") { // This is not settable from outside world, only used for mock tests.
		opts = append(opts, option.WithHTTPClient(&http.Client{Transport: &http.Transport{
			// ignore unverified/expired SSL certificates for tests.
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, //nolint:gosec
		}}))
	}

	setAuthenticationMethod(l, c, opts)

	ctx, cancel := context.WithTimeout(context.Background(), c.Timeout)
	defer cancel()

	client, err := gcstorage.NewClient(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("gcs client initialization, %w", err)
	}

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
	errCh := make(chan error)

	go func() {
		defer close(errCh)

		bkt := b.client.Bucket(b.bucket)
		obj := bkt.Object(p)

		if b.encryption != "" {
			obj = obj.Key([]byte(b.encryption))
		}

		r, err := obj.NewReader(ctx)
		if err != nil {
			errCh <- fmt.Errorf("get the object, %w", err)
			return
		}

		defer internal.CloseWithErrLogf(b.logger, r, "response body, close defer")

		_, err = io.Copy(w, r)
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
	errCh := make(chan error)

	go func() {
		defer close(errCh)

		bkt := b.client.Bucket(b.bucket)
		obj := bkt.Object(p)

		if b.encryption != "" {
			obj = obj.Key([]byte(b.encryption))
		}

		w := obj.NewWriter(ctx)
		defer internal.CloseWithErrLogf(b.logger, w, "object writer, close defer")

		_, err := io.Copy(w, r)
		if err != nil {
			errCh <- fmt.Errorf("copy the object, %w", err)
		}

		if err := w.Close(); err != nil {
			errCh <- fmt.Errorf("close the object, %w", err)
		}

		if b.acl != "" {
			if err := obj.ACL().Set(ctx, gcstorage.AllAuthenticatedUsers, gcstorage.ACLRole(b.acl)); err != nil {
				errCh <- fmt.Errorf("set ACL of the object, %w", err)
			}
		}
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}

// Helpers

func setAuthenticationMethod(l log.Logger, c Config, opts []option.ClientOption) []option.ClientOption {
	if c.APIKey != "" {
		opts = append(opts, option.WithAPIKey(c.APIKey))
		return opts
	}

	creds, err := credentials(l, c)
	if err == nil {
		opts = append(opts, option.WithCredentials(creds))
		return opts
	}

	level.Error(l).Log("msg", "gc storage credential", "err", err)
	level.Warn(l).Log("msg", "initializing gcs without authentication")

	opts = append(opts, option.WithoutAuthentication())

	return opts
}

func credentials(l log.Logger, c Config) (*google.Credentials, error) {
	ctx, cancel := context.WithTimeout(context.Background(), c.Timeout)
	defer cancel()

	creds, err := google.CredentialsFromJSON(ctx, []byte(c.JSONKey), gcstorage.ScopeFullControl)
	if err == nil {
		return creds, nil
	}

	level.Error(l).Log("msg", "gc storage credentials from api-key", "err", err)

	creds, err = google.FindDefaultCredentials(ctx, gcstorage.ScopeFullControl)
	if err != nil {
		return nil, err
	}

	return creds, nil
}
