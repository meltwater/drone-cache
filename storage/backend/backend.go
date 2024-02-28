package backend

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"

	"github.com/meltwater/drone-cache/storage/backend/azure"
	"github.com/meltwater/drone-cache/storage/backend/filesystem"
	"github.com/meltwater/drone-cache/storage/backend/gcs"
	"github.com/meltwater/drone-cache/storage/backend/harness"
	"github.com/meltwater/drone-cache/storage/backend/s3"
	"github.com/meltwater/drone-cache/storage/backend/sftp"
	"github.com/meltwater/drone-cache/storage/common"
)

const (
	// Azure type of the corresponding backend represented as string constant.
	Azure = "azure"
	// FileSystem type of the corresponding backend represented as string constant.
	FileSystem = "filesystem"
	// GCS type of the corresponding backend represented as string constant.
	GCS = "gcs"
	// S3 type of the corresponding backend represented as string constant.
	S3 = "s3"
	// SFTP type of the corresponding backend represented as string constant.
	SFTP = "sftp"
	//Harness type of the corresponding backend represented as string constant.
	Harness = "harness"
)

// Backend implements operations for caching files.
type Backend interface {
	// Get writes downloaded content to the given writer.
	Get(ctx context.Context, p string, w io.Writer) error

	// Put uploads contents of the given reader.
	Put(ctx context.Context, p string, r io.Reader) error

	// Exists checks if path already exists.
	Exists(ctx context.Context, p string) (bool, error)

	// List contents of the given directory by given key from remote storage.
	List(ctx context.Context, p string) ([]common.FileEntry, error)

	// Implement me!
	// Delete(ctx context.Context, p string) error
}

// FromConfig creates new Backend by initializing  using given configuration.
func FromConfig(l log.Logger, backedType string, cfg Config) (Backend, error) {
	var (
		b   Backend
		err error
	)

	switch backedType {
	case Azure:
		level.Debug(l).Log("msg", "using azure blob as backend")
		b, err = azure.New(log.With(l, "backend", Azure), cfg.Azure)
	case S3:
		level.Debug(l).Log("msg", "using aws s3 as backend")
		b, err = s3.New(log.With(l, "backend", S3), cfg.S3, cfg.Debug)
	case Harness:
		level.Debug(l).Log("msg", "using harness as backend")
		b, err = harness.New(log.With(l, "backend", Harness), cfg.Harness, cfg.Debug)
	case GCS:
		level.Debug(l).Log("msg", "using gc storage as backend")
		b, err = gcs.New(log.With(l, "backend", GCS), cfg.GCS)
	case FileSystem:
		level.Debug(l).Log("msg", "using filesystem as backend")
		b, err = filesystem.New(log.With(l, "backend", FileSystem), cfg.FileSystem)
	case SFTP:
		level.Debug(l).Log("msg", "using sftp as backend")
		b, err = sftp.New(log.With(l, "backend", SFTP), cfg.SFTP)
	default:
		return nil, errors.New("unknown backend")
	}

	if err != nil {
		return nil, fmt.Errorf("initialize backend, %w", err)
	}

	return b, nil
}
