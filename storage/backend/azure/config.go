package azure

import "time"

// Config is a structure to store Azure backend configuration
type Config struct {
	AccountName      string
	AccountKey       string
	ContainerName    string
	BlobStorageURL   string
	Azurite          bool
	MaxRetryRequests int
	Timeout          time.Duration
}
