package backend

import (
	"io"

	"github.com/pkg/sftp"
)

type sftpBackend struct {
	client    *sftp.Client
	cacheRoot string
}

func newSftpBackend(client *sftp.Client, cacheRoot string) *sftpBackend {
	return &sftpBackend{client: client, cacheRoot: cacheRoot}
}

func (s sftpBackend) Get(string) (io.ReadCloser, error) {
	panic("implement me")
}

func (s sftpBackend) Put(string, io.ReadSeeker) error {
	panic("implement me")
}
