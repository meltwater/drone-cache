package backend

import (
	"fmt"
	"io"
	"path/filepath"

	"github.com/pkg/sftp"
)

type sftpBackend struct {
	client    *sftp.Client
	cacheRoot string
}

func newSftpBackend(client *sftp.Client, cacheRoot string) *sftpBackend {
	return &sftpBackend{client: client, cacheRoot: cacheRoot}
}

func (s sftpBackend) Get(path string) (io.ReadCloser, error) {
	absPath, err := filepath.Abs(filepath.Clean(filepath.Join(s.cacheRoot, path)))
	if err != nil {
		return nil, fmt.Errorf("get the object %w", err)
	}

	return s.client.Open(absPath)
}

func (s sftpBackend) Put(path string, src io.ReadSeeker) error {
	pathJoin := filepath.Join(s.cacheRoot, path)

	dir := filepath.Dir(pathJoin)
	if err := s.client.MkdirAll(dir); err != nil {
		return fmt.Errorf("create directory <%s> %w", dir, err)
	}

	dst, err := s.client.Create(pathJoin)
	if err != nil {
		return fmt.Errorf("create cache file <%s> %w", pathJoin, err)
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		return fmt.Errorf("write read seeker as file %w", err)
	}

	return nil
}
