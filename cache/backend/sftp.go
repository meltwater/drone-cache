package backend

import (
	"fmt"
	"io"
	"path/filepath"

	"github.com/pkg/errors"
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
		return nil, errors.Wrap(err, "could not get the object")
	}

	return s.client.Open(absPath)
}

func (s sftpBackend) Put(path string, src io.ReadSeeker) error {
	pathJoin := filepath.Join(s.cacheRoot, path)

	dir := filepath.Dir(pathJoin)
	if err := s.client.MkdirAll(dir); err != nil {
		return errors.Wrap(err, fmt.Sprintf("could not create directory <%s>", dir))
	}

	dst, err := s.client.Create(pathJoin)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("could not create cache file <%s>", pathJoin))
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		return errors.Wrap(err, "could not write read seeker as file")
	}

	return nil
}
