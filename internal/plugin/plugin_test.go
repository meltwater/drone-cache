//go:build integration
// +build integration

package plugin

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	gcstorage "cloud.google.com/go/storage"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	awss3 "github.com/aws/aws-sdk-go/service/s3"
	"github.com/go-kit/kit/log"
	pkgsftp "github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
	"google.golang.org/api/option"

	"github.com/meltwater/drone-cache/archive"
	"github.com/meltwater/drone-cache/internal/metadata"
	"github.com/meltwater/drone-cache/storage/backend"
	"github.com/meltwater/drone-cache/storage/backend/azure"
	"github.com/meltwater/drone-cache/storage/backend/filesystem"
	"github.com/meltwater/drone-cache/storage/backend/gcs"
	"github.com/meltwater/drone-cache/storage/backend/s3"
	"github.com/meltwater/drone-cache/storage/backend/sftp"
	"github.com/meltwater/drone-cache/test"
)

const (
	testRoot                       = "testdata"
	testRootMounted                = "testdata/mounted"
	testRootMoved                  = "testdata/moved"
	defaultStorageOperationTimeout = 5 * time.Second
	defaultPublicHost              = "localhost:4443"
	repoName                       = "drone-cache"
)

var publicHost = getEnv("TEST_STORAGE_EMULATOR_HOST", defaultPublicHost)

type setupBackend func(*testing.T, *Config, string)

var (
	backends = map[string]setupBackend{
		// backend.Azure:      setupAzure,
		backend.FileSystem: setupFileSystem,
		backend.GCS:        setupGCS,
		backend.S3:         setupS3,
		// backend.SFTP:       setupSFTP,
	}

	formats = []string{
		archive.Gzip,
		archive.Tar,
		archive.Zstd,
	}
)

func TestPlugin(t *testing.T) {
	test.Ok(t, os.MkdirAll(testRootMounted, 0755))
	test.Ok(t, os.MkdirAll(testRootMoved, 0755))
	t.Cleanup(func() {
		os.RemoveAll(testRoot)
		os.Unsetenv("STORAGE_EMULATOR_HOST") // NOTICE: Only needed for GCS
	})

	cases := []struct {
		name     string
		mount    func(string) []string
		cacheKey string
		success  bool
	}{
		{
			name: "existing mount",
			mount: func(name string) []string {
				return exampleFileTree(t, name, make([]byte, 1*1024))
			},
			success: true,
		},
		{
			name: "non-existing mount",
			mount: func(_ string) []string {
				return []string{"idonotexist"}
			},
			success: false,
		},
		{
			name: "empty mount",
			mount: func(name string) []string {
				return []string{exampleDir(t, name)}
			},
			success: true,
		},
		{
			name: "existing mount with nested files",
			mount: func(name string) []string {
				return exampleNestedFileTree(t, name, make([]byte, 1*1024))
			},
			success: true,
		},
		{
			name: "existing mount with cache key",
			mount: func(name string) []string {
				return exampleFileTree(t, name, make([]byte, 1*1024))
			},
			cacheKey: "{{ .Repo.Name }}_{{ .Commit.Branch }}_{{ .Build.Number }}",
			success:  true,
		},
		{
			name: "existing mount with symlink",
			mount: func(name string) []string {
				return exampleFileTreeWithSymlinks(t, name, make([]byte, 1*1024))
			},
			success: true,
		},
		// NOTICE: Slows down test runs significantly, disabled for now. Will be introduced with a special flag.
		// {
		// 	name: "existing mount with large file",
		// 	mount: func(name string) []string {
		// 		return exampleFileTree(t, "existing", make([]byte, 1*1024*1024))
		// 	},
		// 	success: true,
		// },
	}

	for i, tc := range cases {
		i, tc := i, tc // NOTICE: https://github.com/golang/go/wiki/CommonMistakes#using-goroutines-on-loop-iterator-variables.
		for _, f := range formats {
			for b, setup := range backends {
				name := strings.Join([]string{strconv.Itoa(i), tc.name, b, f}, "-")
				t.Run(name, func(t *testing.T) {
					// Setup
					c := defaultConfig()
					setup(t, c, name)
					paths := tc.mount(tc.name)
					mount(c, paths...)
					cacheKey(c, tc.cacheKey)
					format(c, f)

					// Rebuild run
					{
						plugin := newPlugin(rebuild(c))
						if !tc.success {
							test.NotOk(t, plugin.Exec())
							return
						}

						test.Ok(t, plugin.Exec())
					}

					// Move source to compare later
					restoreRoot, cleanup := test.CreateTempDir(t, sanitize(name), testRootMoved)
					t.Cleanup(cleanup)

					for _, p := range paths {
						rel, err := filepath.Rel(testRootMounted, p)
						test.Ok(t, err)
						dst := filepath.Join(restoreRoot, rel)
						test.Ok(t, os.MkdirAll(filepath.Dir(dst), 0755))
						test.Ok(t, os.Rename(p, dst))
					}

					// Restore run
					{
						if _, ok := os.LookupEnv("STORAGE_EMULATOR_HOST"); !ok { // NOTICE: Only needed for GCS
							test.Ok(t, os.Setenv("STORAGE_EMULATOR_HOST", publicHost))
						}

						plugin := newPlugin(restore(c))
						test.Ok(t, plugin.Exec())

						test.Ok(t, os.Unsetenv("STORAGE_EMULATOR_HOST")) // NOTICE: Only needed for GCS
					}

					// Compare
					test.EqualDirs(t, restoreRoot, testRootMounted, paths)
				})
			}
		}
	}
}

// Plugin configuration

func defaultConfig() *Config {
	return &Config{
		CompressionLevel:        archive.DefaultCompressionLevel,
		StorageOperationTimeout: defaultStorageOperationTimeout,
		Override:                true,
	}
}

func rebuild(c *Config) *Config {
	c.Restore = false
	c.Rebuild = true
	return c
}

func restore(c *Config) *Config {
	c.Restore = true
	c.Rebuild = false
	return c
}

func mount(c *Config, mount ...string) *Config {
	c.Mount = mount
	return c
}

func cacheKey(c *Config, key string) *Config {
	c.CacheKeyTemplate = key
	return c
}

func format(c *Config, fmt string) *Config {
	c.ArchiveFormat = fmt
	return c
}

func newPlugin(c *Config) Plugin {
	var logger log.Logger
	if testing.Verbose() {
		logger = log.NewLogfmtLogger(log.NewSyncWriter(os.Stderr))
	} else {
		logger = log.NewNopLogger()
	}

	return Plugin{
		logger: logger,
		Metadata: metadata.Metadata{
			Repo: metadata.Repo{
				Branch: "master",
				Name:   repoName,
			},
			Commit: metadata.Commit{
				Branch: "master",
			},
		},
		Config: *c,
	}
}

// Fixtures

func exampleDir(t *testing.T, name string) string {
	name = sanitize(name)

	dir, cleanup := test.CreateTempDir(t, name, testRootMounted)
	t.Cleanup(cleanup)

	return dir
}

func exampleFileTree(t *testing.T, name string, content []byte) []string {
	name = sanitize(name)

	file, fileClean := test.CreateTempFile(t, name, content, testRootMounted)
	t.Cleanup(fileClean)

	dir, dirClean := test.CreateTempFilesInDir(t, name, content, testRootMounted)
	t.Cleanup(dirClean)

	return []string{file, dir}
}

func exampleNestedFileTree(t *testing.T, name string, content []byte) []string {
	name = sanitize(name)

	dir, cleanup := test.CreateTempDir(t, name, testRootMounted)
	t.Cleanup(cleanup)

	nestedFile, nestedFileClean := test.CreateTempFile(t, name, content, dir)
	t.Cleanup(nestedFileClean)

	nestedDir, nestedDirClean := test.CreateTempFilesInDir(t, name, content, dir)
	t.Cleanup(nestedDirClean)

	nestedDir1, nestedDirClean1 := test.CreateTempDir(t, name, dir)
	t.Cleanup(nestedDirClean1)

	nestedFile1, nestedFileClean1 := test.CreateTempFile(t, name, content, nestedDir1)
	t.Cleanup(nestedFileClean1)

	return []string{nestedDir, nestedFile, nestedFile1}
}

func exampleFileTreeWithSymlinks(t *testing.T, name string, content []byte) []string {
	name = sanitize(name)

	file, fileClean := test.CreateTempFile(t, name, content, testRootMounted)
	t.Cleanup(fileClean)

	dir, dirClean := test.CreateTempFilesInDir(t, name, content, testRootMounted)
	t.Cleanup(dirClean)

	symDir, cleanup := test.CreateTempDir(t, name, testRootMounted)
	t.Cleanup(cleanup)

	symlink := filepath.Join(symDir, name+"_symlink.testfile")
	test.Ok(t, os.Symlink(file, symlink))
	t.Cleanup(func() { os.Remove(symlink) })

	return []string{file, dir, symDir}
}

// Setup

func setupAzure(t *testing.T, c *Config, name string) {
	const (
		defaultBlobStorageURL = "127.0.0.1:10000"
		defaultAccountName    = "devstoreaccount1"
		defaultAccountKey     = "Eby8vdM02xNOcqFlqUwJPLlmEtlCDXJ1OUzFT50uSRZ6IFsuFq2UVErCz4I6tq/K1SZFPTOtr/KBHBeksoGMGw=="
	)

	var (
		blobURL     = getEnv("TEST_AZURITE_URL", defaultBlobStorageURL)
		accountName = getEnv("TEST_ACCOUNT_NAME", defaultAccountName)
		accountKey  = getEnv("TEST_ACCOUNT_KEY", defaultAccountKey)
	)

	c.Backend = backend.Azure
	c.Azure = azure.Config{
		AccountName:    accountName,
		AccountKey:     accountKey,
		ContainerName:  name,
		BlobStorageURL: blobURL,
		Azurite:        true,
		Timeout:        defaultStorageOperationTimeout,
	}
}

func setupFileSystem(t *testing.T, c *Config, name string) {
	dir, cleanup := test.CreateTempDir(t, "filesystem-cache-root-"+sanitize(name), "testdata")
	t.Cleanup(cleanup)

	c.Backend = backend.FileSystem
	c.FileSystem = filesystem.Config{CacheRoot: dir}
}

func setupGCS(t *testing.T, c *Config, name string) {
	const (
		defaultEndpoint = "http://127.0.0.1:4443/storage/v1/"
		defaultApiKey   = ""
	)

	var (
		endpoint   = getEnv("TEST_GCS_ENDPOINT", defaultEndpoint)
		apiKey     = getEnv("TEST_GCS_API_KEY", defaultApiKey)
		bucketName = sanitize(name)
		opts       []option.ClientOption
	)

	if apiKey != "" {
		opts = append(opts, option.WithAPIKey(apiKey))
	} else {
		opts = append(opts, option.WithoutAuthentication())
	}
	opts = append(opts, option.WithEndpoint(endpoint))
	opts = append(opts, option.WithHTTPClient(&http.Client{Transport: &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, // ignore expired SSL certificates.
	}}))

	client, err := gcstorage.NewClient(context.Background(), opts...)
	test.Ok(t, err)

	bucket := client.Bucket(bucketName)
	test.Ok(t, bucket.Create(context.Background(), "drone-cache", &gcstorage.BucketAttrs{}))
	t.Cleanup(func() { client.Close() })

	c.Backend = backend.GCS
	c.GCS = gcs.Config{
		Bucket:   bucketName,
		Endpoint: endpoint,
		APIKey:   apiKey,
		Timeout:  defaultStorageOperationTimeout,
	}
}

func setupS3(t *testing.T, c *Config, name string) {
	const (
		defaultEndpoint        = "127.0.0.1:9000"
		defaultAccessKey       = "AKIAIOSFODNN7EXAMPLE"
		defaultSecretAccessKey = "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY"
		defaultRegion          = "eu-west-1"
	)
	var (
		endpoint        = getEnv("TEST_S3_ENDPOINT", defaultEndpoint)
		accessKey       = getEnv("TEST_S3_ACCESS_KEY", defaultAccessKey)
		secretAccessKey = getEnv("TEST_S3_SECRET_KEY", defaultSecretAccessKey)
		bucket          = sanitize(name)
	)
	client := awss3.New(session.Must(session.NewSessionWithOptions(session.Options{})), &aws.Config{
		Region:           aws.String(defaultRegion),
		Endpoint:         aws.String(endpoint),
		DisableSSL:       aws.Bool(!strings.HasPrefix(endpoint, "https://")),
		S3ForcePathStyle: aws.Bool(true),
		Credentials:      credentials.NewStaticCredentials(accessKey, secretAccessKey, ""),
	})

	_, err := client.CreateBucketWithContext(context.Background(), &awss3.CreateBucketInput{
		Bucket: aws.String(bucket),
	})
	test.Ok(t, err)

	c.Backend = backend.S3
	c.S3 = s3.Config{
		ACL:       "private",
		Bucket:    bucket,
		Endpoint:  endpoint,
		Key:       accessKey,
		PathStyle: true, // Should be true for minio and false for AWS.
		Region:    defaultRegion,
		Secret:    secretAccessKey,
	}
}

func setupSFTP(t *testing.T, c *Config, name string) {
	const (
		defaultSFTPHost  = "127.0.0.1"
		defaultSFTPPort  = "22"
		defaultUsername  = "bar"
		defaultPassword  = "pass"
		defaultCacheRoot = "/plugin_test"
	)

	var (
		host      = getEnv("TEST_SFTP_HOST", defaultSFTPHost)
		port      = getEnv("TEST_SFTP_PORT", defaultSFTPPort)
		username  = getEnv("TEST_SFTP_USERNAME", defaultUsername)
		password  = getEnv("TEST_SFTP_PASSWORD", defaultPassword)
		cacheRoot = filepath.Join(getEnv("TEST_SFTP_CACHE_ROOT", defaultCacheRoot), "sft-cache-root-"+sanitize(name))
	)

	/* #nosec */
	sshClient, err := ssh.Dial("tcp", fmt.Sprintf("%s:%s", host, port), &ssh.ClientConfig{
		User:            username,
		Auth:            []ssh.AuthMethod{ssh.Password(password)},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // #nosec TODO(kakkoyun) just a workaround for now, will fix
	})
	test.Ok(t, err)

	client, err := pkgsftp.NewClient(sshClient)
	test.Ok(t, err)

	test.Ok(t, client.MkdirAll(filepath.Join(cacheRoot, repoName)))

	c.Backend = backend.SFTP
	c.SFTP = sftp.Config{
		CacheRoot: cacheRoot,
		Username:  username,
		Auth: sftp.SSHAuth{
			Password: password,
			Method:   sftp.SSHAuthMethodPassword,
		},
		Host: host,
		Port: port,
	}
}

// Helpers

func sanitize(p string) string {
	return strings.ReplaceAll(strings.TrimSpace(strings.ToLower(p)), " ", "-")
}

func getEnv(key, defaultVal string) string {
	value, ok := os.LookupEnv(key)
	if !ok {
		return defaultVal
	}
	return value
}
