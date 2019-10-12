package backend

import (
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"path"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/meltwater/drone-cache/cache"
	"github.com/pkg/errors"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

// S3Config is a structure to store S3  backend configuration
type S3Config struct {
	// Indicates the files ACL, which should be one
	// of the following:
	//     private
	//     public-read
	//     public-read-write
	//     authenticated-read
	//     bucket-owner-read
	//     bucket-owner-full-control
	ACL        string
	Bucket     string
	Encryption string // if not "", enables server-side encryption. valid values are: AES256, aws:kms
	Endpoint   string
	Key        string

	// us-east-1
	// us-west-1
	// us-west-2
	// eu-west-1
	// ap-southeast-1
	// ap-southeast-2
	// ap-northeast-1
	// sa-east-1
	Region string
	Secret string

	PathStyle bool // Use path style instead of domain style. Should be true for minio and false for AWS
}

// FileSystemConfig is a structure to store filesystem backend configuration
type FileSystemConfig struct {
	CacheRoot string
}

// InitializeS3Backend creates an S3 backend
func InitializeS3Backend(c S3Config, debug bool) (cache.Backend, error) {
	awsConf := &aws.Config{
		Region:           aws.String(c.Region),
		Endpoint:         &c.Endpoint,
		DisableSSL:       aws.Bool(!strings.HasPrefix(c.Endpoint, "https://")),
		S3ForcePathStyle: aws.Bool(c.PathStyle),
	}

	if c.Key != "" && c.Secret != "" {
		awsConf.Credentials = credentials.NewStaticCredentials(c.Key, c.Secret, "")
	} else {
		log.Println("aws key and/or Secret not provided (falling back to anonymous credentials)")
	}

	if debug {
		log.Printf("[DEBUG] s3 backend config: %+v", c)
		awsConf.WithLogLevel(aws.LogDebugWithHTTPBody)
	}

	return newS3(c.Bucket, c.ACL, c.Encryption, awsConf), nil
}

// InitializeFileSystemBackend creates a filesystem backend
func InitializeFileSystemBackend(c FileSystemConfig, debug bool) (cache.Backend, error) {
	if strings.TrimRight(path.Clean(c.CacheRoot), "/") == "" {
		return nil, fmt.Errorf("could not use <%s> as cache root, empty or root path given", c.CacheRoot)
	}

	if _, err := os.Stat(c.CacheRoot); err != nil {
		msg := fmt.Sprintf("could not use <%s> as cache root, make sure volume is mounted", c.CacheRoot)
		return nil, errors.Wrap(err, msg)
	}

	if debug {
		log.Printf("[DEBUG] filesystem backend config: %+v", c)
	}

	return newFileSystem(c.CacheRoot), nil
}

type SSHAuthMethod string

const (
	SSHAuthMethodPassword      SSHAuthMethod = "PASSWORD"
	SSHAuthMethodPublicKeyFile SSHAuthMethod = "PUBLIC_KEY_FILE"
)

type SSHAuth struct {
	Password      string
	PublicKeyFile string
	Method        SSHAuthMethod
}

// SFTPConfig is a structure to store sftp backend configuration
type SFTPConfig struct {
	CacheRoot string
	Username  string
	Host      string
	Port      string
	Auth      SSHAuth
}

func InitializeSFTPBackend(c SFTPConfig, debug bool) (cache.Backend, error) {
	sshClient, err := getSSHClient(c)
	if err != nil {
		return nil, err
	}

	sftpClient, err := sftp.NewClient(sshClient)
	if err != nil {
		return nil, errors.Wrap(err, "unable to connect to ssh with sftp protocol")
	}

	if debug {
		log.Printf("[DEBUG] sftp backend config: %+v", c)
	}

	return newSftpBackend(sftpClient, c.CacheRoot), nil
}

func getSSHClient(c SFTPConfig) (*ssh.Client, error) {
	authMethod, err := getAuthMethod(c)
	if err != nil {
		return nil, errors.Wrap(err, " unable to get ssh auth method")
	}

	client, err := ssh.Dial("tcp", fmt.Sprintf("%s:%s", c.Host, c.Port), &ssh.ClientConfig{
		User: c.Username,
		Auth: authMethod,
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			return nil
		},
	})
	if err != nil {
		return nil, errors.Wrap(err, "unable to connect to ssh")
	}

	return client, nil
}

func getAuthMethod(c SFTPConfig) ([]ssh.AuthMethod, error) {
	if c.Auth.Method == SSHAuthMethodPassword {
		return []ssh.AuthMethod{
			ssh.Password(c.Auth.Password),
		}, nil
	} else if c.Auth.Method == SSHAuthMethodPublicKeyFile {
		pkAuthMethod, err := readPublicKeyFile(c.Auth.PublicKeyFile)
		return []ssh.AuthMethod{
			pkAuthMethod,
		}, err
	}
	return nil, errors.New("ssh method auth is not recognized, should be PASSWORD or PUBLIC_KEY_FILE")
}

func readPublicKeyFile(file string) (ssh.AuthMethod, error) {
	buffer, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("unable to read file <%s>", file))
	}

	key, err := ssh.ParsePrivateKey(buffer)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("unable to parse private key"))
	}
	return ssh.PublicKeys(key), nil
}
