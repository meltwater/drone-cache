package sftp

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"

	"github.com/meltwater/drone-cache/internal"
	"github.com/meltwater/drone-cache/storage/common"
)

// Backend implements storage.Backend for sFTP.
type Backend struct {
	logger log.Logger

	cacheRoot string
	client    *sftp.Client
}

// New creates a new sFTP backend.
func New(l log.Logger, c Config) (*Backend, error) {
	authMethod, err := authMethod(c)
	if err != nil {
		return nil, fmt.Errorf("unable to get ssh auth method, %w", err)
	}

	/* #nosec */
	sshClient, err := ssh.Dial("tcp", fmt.Sprintf("%s:%s", c.Host, c.Port), &ssh.ClientConfig{
		User:            c.Username,
		Auth:            authMethod,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // #nosec TODO(kakkoyun) just a workaround for now, will fix
		Timeout:         c.Timeout,
	})
	if err != nil {
		return nil, fmt.Errorf("unable to connect to ssh, %w", err)
	}

	client, err := sftp.NewClient(sshClient)
	if err != nil {
		sshClient.Close()
		return nil, fmt.Errorf("unable to connect to ssh with sftp protocol, %w", err)
	}

	if _, err := client.Stat(c.CacheRoot); err != nil {
		return nil, fmt.Errorf("make sure cache root <%s> created, %w", c.CacheRoot, err)
	}

	level.Debug(l).Log("msg", "sftp backend", "config", fmt.Sprintf("%#v", c))

	return &Backend{logger: l, client: client, cacheRoot: c.CacheRoot}, nil
}

// Get writes downloaded content to the given writer.
func (b *Backend) Get(ctx context.Context, p string, w io.Writer) error {
	path, err := filepath.Abs(filepath.Clean(filepath.Join(b.cacheRoot, p)))
	if err != nil {
		return fmt.Errorf("generate absolute path, %w", err)
	}

	errCh := make(chan error)

	go func() {
		defer close(errCh)

		rc, err := b.client.Open(path)
		if err != nil {
			errCh <- fmt.Errorf("get the object, %w", err)
			return
		}

		defer internal.CloseWithErrLogf(b.logger, rc, "reader close defer")

		_, err = io.Copy(w, rc)
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

		path := filepath.Clean(filepath.Join(b.cacheRoot, p))

		dir := filepath.Dir(path)
		if err := b.client.MkdirAll(dir); err != nil {
			errCh <- fmt.Errorf("create directory, %w", err)
			return
		}

		w, err := b.client.Create(path)
		if err != nil {
			errCh <- fmt.Errorf("create cache file, %w", err)
			return
		}

		defer internal.CloseWithErrLogf(b.logger, w, "writer close defer")

		if _, err := io.Copy(w, r); err != nil {
			errCh <- fmt.Errorf("write contents of reader to a file, %w", err)
		}

		if err := w.Close(); err != nil {
			errCh <- fmt.Errorf("close the object, %w", err)
		}
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}

// Exists checks if object already exists.
func (b *Backend) Exists(ctx context.Context, p string) (bool, error) {
	path, err := filepath.Abs(filepath.Clean(filepath.Join(b.cacheRoot, p)))
	if err != nil {
		return false, fmt.Errorf("generate absolute path, %w", err)
	}

	type result struct {
		val bool
		err error
	}

	resCh := make(chan *result)

	go func() {
		defer close(resCh)

		_, err := b.client.Stat(path)
		if err != nil && !os.IsNotExist(err) {
			resCh <- &result{err: fmt.Errorf("check the object exists, %w", err)}
			return
		}
		resCh <- &result{val: err == nil}
	}()

	select {
	case res := <-resCh:
		return res.val, res.err
	case <-ctx.Done():
		return false, ctx.Err()
	}
}

// List contents of the given directory by given key from remote storage.
func (b *Backend) List(ctx context.Context, p string) ([]common.FileEntry, error) {
	return nil, common.ErrNotImplemented
}

// Helpers

func authMethod(c Config) ([]ssh.AuthMethod, error) {
	switch c.Auth.Method {
	case SSHAuthMethodPassword:
		return []ssh.AuthMethod{ssh.Password(c.Auth.Password)}, nil
	case SSHAuthMethodPublicKeyFile:
		pkAuthMethod, err := readPublicKeyFile(c.Auth.PublicKeyFile)
		return []ssh.AuthMethod{pkAuthMethod}, err
	default:
		return nil, errors.New("unknown ssh method (PASSWORD, PUBLIC_KEY_FILE)")
	}
}

func readPublicKeyFile(file string) (ssh.AuthMethod, error) {
	buffer, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, fmt.Errorf("unable to read file, %w", err)
	}

	key, err := ssh.ParsePrivateKey(buffer)
	if err != nil {
		return nil, fmt.Errorf("unable to parse private key, %w", err)
	}

	return ssh.PublicKeys(key), nil
}
