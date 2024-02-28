package harness

import (
	"context"
	"fmt"
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
	GetUploadPresignURL(ctx context.Context, key string) (string, error)

	GetDownloadPresignURL(ctx context.Context, key string) (string, error)

	GetExistsPresignURL(ctx context.Context, key string) (string, error)

	GetListPresignURL(ctx context.Context, key string) (string, error)
}
