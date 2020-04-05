package sftp

import "time"

// SSHAuthMethod describes the type of authentication method.
type SSHAuthMethod string

const (
	SSHAuthMethodPassword      SSHAuthMethod = "PASSWORD"
	SSHAuthMethodPublicKeyFile SSHAuthMethod = "PUBLIC_KEY_FILE"
)

// SSHAuth is a structure to store authentication information for SSH connection.
type SSHAuth struct {
	Password      string
	PublicKeyFile string
	Method        SSHAuthMethod
}

// Config is a structure to store sFTP backend configuration
type Config struct {
	CacheRoot string
	Username  string
	Host      string
	Port      string
	Auth      SSHAuth
	Timeout   time.Duration
}
