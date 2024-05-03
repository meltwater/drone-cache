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
	GetUploadURL(ctx context.Context, key string) (string, error)

	GetDownloadURL(ctx context.Context, key string) (string, error)

	GetExistsURL(ctx context.Context, key string) (string, error)

	GetListURL(ctx context.Context, key, continuationToken string) (string, error)
}
