package common

import (
	"time"
)

// FileEntry defines a single cache item.
type FileEntry struct {
	Path         string
	Size         int64
	LastModified time.Time
}
