package gcs

import "time"

// Config is a structure to store Cloud Storage backend configuration.
type Config struct {
	Bucket              string
	ACL                 string
	Encryption          string
	Endpoint            string
	APIKey              string
	JSONKey             string
	Timeout             time.Duration
	OIDCTokenID         string // OIDC releated fields
	ProjectNumber       string
	PoolID              string
	ProviderID          string
	ServiceAccountEmail string
}
