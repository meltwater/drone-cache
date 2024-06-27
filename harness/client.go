package harness

import (
	"context"
	"fmt"
	"github.com/meltwater/drone-cache/storage/common"
)

// Error is a custom error struct
type Error struct {
	Code    int
	Message string
}

func (e *Error) Error() string {
	return fmt.Sprintf("%d: %s", e.Code, e.Message)
}

// Client defines a cache service client.
type Client interface {
	GetUploadURL(ctx context.Context, key string) (string, error)

	GetDownloadURL(ctx context.Context, key string) (string, error)

	GetExistsURL(ctx context.Context, key string) (string, error)

	GetEntriesList(ctx context.Context, prefix string) ([]common.FileEntry, error)
}
