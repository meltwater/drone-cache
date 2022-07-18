package plugin

import (
	"fmt"
	"strings"
	"time"

	"github.com/bmatcuk/doublestar"
	"github.com/meltwater/drone-cache/storage/backend/azure"
	"github.com/meltwater/drone-cache/storage/backend/filesystem"
	"github.com/meltwater/drone-cache/storage/backend/gcs"
	"github.com/meltwater/drone-cache/storage/backend/s3"
	"github.com/meltwater/drone-cache/storage/backend/sftp"
)

// Config plugin-specific parameters and secrets.
type Config struct {
	ArchiveFormat    string
	Backend          string
	CacheKeyTemplate string
	RemoteRoot       string
	LocalRoot        string

	// Modes
	Debug   bool
	Rebuild bool
	Restore bool

	// Optional
	SkipSymlinks            bool
	Override                bool
	CompressionLevel        int
	StorageOperationTimeout time.Duration

	Mount []string

	// Backend
	S3         s3.Config
	FileSystem filesystem.Config
	SFTP       sftp.Config
	Azure      azure.Config
	GCS        gcs.Config
}

func (c *Config) HandleMount() error {
	mountLen := len(c.Mount)
	if mountLen > 0 {
		for i, mount := range c.Mount {
			if strings.Contains(mount, "**") {
				// Remove the glob from the original mount list
				c.Mount[i] = c.Mount[mountLen-1]
				c.Mount = c.Mount[:mountLen-1]

				globMounts, err := doublestar.Glob(mount)
				if err != nil {
					return fmt.Errorf("glob handle mount error <%s>, %w", mount, err)
				}

				c.Mount = append(c.Mount, globMounts...)
			}
		}
	}

	return nil
}
