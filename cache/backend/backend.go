package backend

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"path"
	"strings"

	"github.com/meltwater/drone-cache/cache"

	"github.com/Azure/azure-storage-blob-go/azblob"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
	"google.golang.org/api/option"
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

// AzureConfig is a structure to store Azure backend configuration
type AzureConfig struct {
	AccountName    string
	AccountKey     string
	ContainerName  string
	BlobStorageURL string
	Azurite        bool
}

// FileSystemConfig is a structure to store filesystem backend configuration
type FileSystemConfig struct {
	CacheRoot string
}

// InitializeS3Backend creates an S3 backend
func InitializeS3Backend(l log.Logger, c S3Config, debug bool) (cache.Backend, error) {
	awsConf := &aws.Config{
		Region:           aws.String(c.Region),
		Endpoint:         &c.Endpoint,
		DisableSSL:       aws.Bool(!strings.HasPrefix(c.Endpoint, "https://")),
		S3ForcePathStyle: aws.Bool(c.PathStyle),
	}

	if c.Key != "" && c.Secret != "" {
		awsConf.Credentials = credentials.NewStaticCredentials(c.Key, c.Secret, "")
	} else {
		level.Warn(l).Log("msg", "aws key and/or Secret not provided (falling back to anonymous credentials)")
	}

	level.Debug(l).Log("msg", "s3 backend", "config", fmt.Sprintf("%+v", c))

	if debug {
		awsConf.WithLogLevel(aws.LogDebugWithHTTPBody)
	}

	return newS3(c.Bucket, c.ACL, c.Encryption, awsConf), nil
}

// InitializeAzureBackend creates an AzureBlob backend
func InitializeAzureBackend(l log.Logger, c AzureConfig, debug bool) (cache.Backend, error) {
	// From the Azure portal, get your storage account name and key and set environment variables.
	accountName, accountKey := c.AccountName, c.AccountKey
	if len(accountName) == 0 || len(accountKey) == 0 {
		return nil, fmt.Errorf("either the AZURE_ACCOUNT_NAME or AZURE_ACCOUNT_KEY environment variable is not set")
	}

	// Create a default request pipeline using your storage account name and account key.
	credential, err := azblob.NewSharedKeyCredential(accountName, accountKey)
	if err != nil {
		level.Error(l).Log("msg", "invalid credentials with error: "+err.Error())
	}

	var azureBlobURL *url.URL

	// Azurite has different URL pattern than production Azure Blob Storage
	if c.Azurite {
		azureBlobURL, err = url.Parse(fmt.Sprintf("http://%s/%s/%s", c.BlobStorageURL, c.AccountName, c.ContainerName))
	} else {
		azureBlobURL, err = url.Parse(fmt.Sprintf("https://%s.%s/%s", c.AccountName, c.BlobStorageURL, c.ContainerName))
	}

	if err != nil {
		level.Error(l).Log("msg", "can't create url with : "+err.Error())
	}

	pipeline := azblob.NewPipeline(credential, azblob.PipelineOptions{})
	containerURL := azblob.NewContainerURL(*azureBlobURL, pipeline)
	ctx := context.Background()

	// Always creating new container, it will throw error if it already exists
	_, err = containerURL.Create(ctx, azblob.Metadata{}, azblob.PublicAccessNone)
	if err != nil {
		level.Debug(l).Log("msg", "container already exists:"+err.Error())
	}

	return newAzure(ctx, containerURL), nil
}

// InitializeFileSystemBackend creates a filesystem backend
func InitializeFileSystemBackend(l log.Logger, c FileSystemConfig, debug bool) (cache.Backend, error) {
	if strings.TrimRight(path.Clean(c.CacheRoot), "/") == "" {
		return nil, fmt.Errorf("empty or root path given, <%s> as cache root, ", c.CacheRoot)
	}

	if _, err := os.Stat(c.CacheRoot); err != nil {
		return nil, fmt.Errorf("make sure volume is mounted, <%s> as cache root %w", c.CacheRoot, err)
	}

	level.Debug(l).Log("msg", "filesystem backend", "config", fmt.Sprintf("%+v", c))

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

func InitializeSFTPBackend(l log.Logger, c SFTPConfig, debug bool) (cache.Backend, error) {
	sshClient, err := getSSHClient(c)
	if err != nil {
		return nil, err
	}

	sftpClient, err := sftp.NewClient(sshClient)
	if err != nil {
		return nil, fmt.Errorf("unable to connect to ssh with sftp protocol %w", err)
	}

	level.Debug(l).Log("msg", "sftp backend", "config", fmt.Sprintf("%+v", c))

	return newSftpBackend(sftpClient, c.CacheRoot), nil
}

func getSSHClient(c SFTPConfig) (*ssh.Client, error) {
	authMethod, err := getAuthMethod(c)
	if err != nil {
		return nil, fmt.Errorf("unable to get ssh auth method %w", err)
	}

	/* #nosec */
	client, err := ssh.Dial("tcp", fmt.Sprintf("%s:%s", c.Host, c.Port), &ssh.ClientConfig{
		User:            c.Username,
		Auth:            authMethod,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // #nosec just a workaround for now, will fix
	})
	if err != nil {
		return nil, fmt.Errorf("unable to connect to ssh %w", err)
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
		return nil, fmt.Errorf("unable to read file <%s> %w", file, err)
	}

	key, err := ssh.ParsePrivateKey(buffer)
	if err != nil {
		return nil, fmt.Errorf("unable to parse private key %w", err)
	}

	return ssh.PublicKeys(key), nil
}

// CloudStorageConfig is a structure to store Cloud Storage backend configuration
type CloudStorageConfig struct {
	Bucket     string
	ACL        string
	Encryption string
	Endpoint   string
	APIKey     string
}

// InitializeGCSBackend creates a Cloud Storage backend
func InitializeGCSBackend(l log.Logger, c CloudStorageConfig, debug bool) (cache.Backend, error) {
	var opts []option.ClientOption
	if c.APIKey != "" {
		opts = append(opts, option.WithAPIKey(c.APIKey))
	}

	if c.Endpoint != "" {
		opts = append(opts, option.WithEndpoint(c.Endpoint))
	}

	if debug {
		level.Debug(l).Log("msg", "gc storage backend", "config", fmt.Sprintf("%+v", c))
	}

	return newGCS(c.Bucket, c.ACL, c.Encryption, opts...)
}
