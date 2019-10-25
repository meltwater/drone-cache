package plugin

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/meltwater/drone-cache/cache"
	"github.com/meltwater/drone-cache/cache/backend"
	"github.com/meltwater/drone-cache/metadata"

	"github.com/go-kit/kit/log"
	"github.com/minio/minio-go"
)

const (
	defaultEndpoint        = "127.0.0.1:9000"
	defaultAccessKey       = "AKIAIOSFODNN7EXAMPLE"
	defaultSecretAccessKey = "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY"
	bucket                 = "meltwater-drone-test"
	region                 = "eu-west-1"
	useSSL                 = false
)

var (
	endpoint        = getEnv("TEST_ENDPOINT", defaultEndpoint)
	accessKey       = getEnv("TEST_ACCESS_KEY", defaultAccessKey)
	secretAccessKey = getEnv("TEST_SECRET_KEY", defaultSecretAccessKey)
)

func TestRebuild(t *testing.T) {
	setup(t)
	defer cleanUp(t)

	dirPath := "./tmp/1"
	if mkErr1 := os.MkdirAll(dirPath, 0755); mkErr1 != nil {
		t.Fatal(mkErr1)
	}

	fPath := "./tmp/1/file_to_cache.txt"
	file, fErr := os.Create(fPath)
	if fErr != nil {
		t.Fatal(fErr)
	}

	if _, err := file.WriteString("some content\n"); err != nil {
		t.Fatal(err)
	}
	file.Sync()
	file.Close()

	absPath, err := filepath.Abs(fPath)
	if err != nil {
		t.Fatal(err)
	}

	linkAbsPath, err := filepath.Abs("./tmp/1/symlink_to_cache.txt")
	if err != nil {
		t.Fatal(err)
	}

	if err := os.Symlink(absPath, linkAbsPath); err != nil {
		t.Fatal(err)
	}

	plugin := newTestPlugin("s3", true, false, []string{dirPath}, "", "")

	if err := plugin.Exec(); err != nil {
		t.Errorf("plugin exec failed, error: %v\n", err)
	}
}

func TestRebuildSkipSymlinks(t *testing.T) {
	setup(t)
	defer cleanUp(t)

	dirPath := "./tmp/1"
	if mkErr1 := os.MkdirAll(dirPath, 0755); mkErr1 != nil {
		t.Fatal(mkErr1)
	}

	fPath := "./tmp/1/file_to_cache.txt"
	file, fErr := os.Create(fPath)
	if fErr != nil {
		t.Fatal(fErr)
	}

	if _, err := file.WriteString("some content\n"); err != nil {
		t.Fatal(err)
	}
	file.Sync()
	file.Close()

	absPath, err := filepath.Abs(fPath)
	if err != nil {
		t.Fatal(err)
	}

	linkAbsPath, err := filepath.Abs("./tmp/1/symlink_to_cache.txt")
	if err != nil {
		t.Fatal(err)
	}

	if err := os.Symlink(absPath, linkAbsPath); err != nil {
		t.Fatal(err)
	}

	plugin := newTestPlugin("s3", true, false, []string{"./tmp/1"}, "", "")
	plugin.Config.SkipSymlinks = true

	if err := plugin.Exec(); err != nil {
		t.Errorf("plugin exec failed, error: %v\n", err)
	}
}

func TestRebuildWithCacheKey(t *testing.T) {
	setup(t)
	defer cleanUp(t)

	if mkErr1 := os.MkdirAll("./tmp/1", 0755); mkErr1 != nil {
		t.Fatal(mkErr1)
	}

	file, fErr := os.Create("./tmp/1/file_to_cache.txt")
	if fErr != nil {
		t.Fatal(fErr)
	}

	if _, err := file.WriteString("some content\n"); err != nil {
		t.Fatal(err)
	}
	file.Sync()
	file.Close()

	plugin := newTestPlugin("s3", true, false, []string{"./tmp/1"}, "{{ .Repo.Name }}_{{ .Commit.Branch }}_{{ .Build.Number }}", "")

	if err := plugin.Exec(); err != nil {
		t.Errorf("plugin exec failed, error: %v\n", err)
	}
}

func TestRebuildWithGzip(t *testing.T) {
	setup(t)
	defer cleanUp(t)

	if mkErr1 := os.MkdirAll("./tmp/1", 0755); mkErr1 != nil {
		t.Fatal(mkErr1)
	}

	file, fErr := os.Create("./tmp/1/file_to_cache.txt")
	if fErr != nil {
		t.Fatal(fErr)
	}

	if _, err := file.WriteString("some content\n"); err != nil {
		t.Fatal(err)
	}
	file.Sync()
	file.Close()

	plugin := newTestPlugin("s3", true, false, []string{"./tmp/1"}, "", "gzip")

	if err := plugin.Exec(); err != nil {
		t.Errorf("plugin exec failed, error: %v\n", err)
	}
}

func TestRebuildWithFilesystem(t *testing.T) {
	setup(t)
	defer cleanUp(t)

	if mkErr1 := os.MkdirAll("./tmp/1", 0755); mkErr1 != nil {
		t.Fatal(mkErr1)
	}

	file, fErr := os.Create("./tmp/1/file_to_cache.txt")
	if fErr != nil {
		t.Fatal(fErr)
	}

	if _, err := file.WriteString("some content\n"); err != nil {
		t.Fatal(err)
	}
	file.Sync()
	file.Close()

	plugin := newTestPlugin("filesystem", true, false, []string{"./tmp/1"}, "", "gzip")

	if err := plugin.Exec(); err != nil {
		t.Errorf("plugin exec failed, error: %v\n", err)
	}
}

func TestRebuildNonExisting(t *testing.T) {
	setup(t)
	defer cleanUp(t)

	plugin := newTestPlugin("s3", true, false, []string{"./nonexisting/path"}, "", "")

	if err := plugin.Exec(); err == nil {
		t.Error("plugin exec did not fail as expected, error: <nil>")
	}
}

func TestRestore(t *testing.T) {
	setup(t)
	defer cleanUp(t)

	dirPath := "./tmp/1"
	if err := os.MkdirAll(dirPath, 0755); err != nil {
		t.Fatal(err)
	}

	if err := os.MkdirAll("./tmp/2", 0755); err != nil {
		t.Fatal(err)
	}

	fPath := "./tmp/1/file_to_cache.txt"
	file, cErr := os.Create(fPath)
	if cErr != nil {
		t.Fatal(cErr)
	}

	if _, err := file.WriteString("some content\n"); err != nil {
		t.Fatal(err)
	}

	file.Sync()
	file.Close()

	file1, fErr1 := os.Create("./tmp/1/file1_to_cache.txt")
	if fErr1 != nil {
		t.Fatal(fErr1)
	}

	if _, err := file1.WriteString("some content\n"); err != nil {
		t.Fatal(err)
	}

	file1.Sync()
	file1.Close()

	absPath, err := filepath.Abs(fPath)
	if err != nil {
		t.Fatal(err)
	}

	linkAbsPath, err := filepath.Abs("./tmp/1/symlink_to_cache.txt")
	if err != nil {
		t.Fatal(err)
	}

	if err := os.Symlink(absPath, linkAbsPath); err != nil {
		t.Fatal(err)
	}

	plugin := newTestPlugin("s3", true, false, []string{dirPath}, "", "")

	if err := plugin.Exec(); err != nil {
		t.Errorf("plugin (rebuild mode) exec failed, error: %v\n", err)
	}

	if err := os.RemoveAll("./tmp"); err != nil {
		t.Fatal(err)
	}

	plugin.Config.Rebuild = false
	plugin.Config.Restore = true
	if err := plugin.Exec(); err != nil {
		t.Errorf("plugin (restore mode) exec failed, error: %v\n", err)
	}

	if _, err := os.Stat("./tmp/1/file_to_cache.txt"); os.IsNotExist(err) {
		t.Error(err)
	}

	if _, err := os.Stat("./tmp/1/file1_to_cache.txt"); os.IsNotExist(err) {
		t.Error(err)
	}

	target, err := os.Readlink("./tmp/1/symlink_to_cache.txt")
	if err != nil {
		t.Error(err)
	}

	if _, err := os.Stat(target); os.IsNotExist(err) {
		t.Error(err)
	}
}

func TestRestoreWithCacheKey(t *testing.T) {
	setup(t)
	defer cleanUp(t)

	if err := os.MkdirAll("./tmp/1", 0755); err != nil {
		t.Fatal(err)
	}

	file, cErr := os.Create("./tmp/1/file_to_cache.txt")
	if cErr != nil {
		t.Fatal(cErr)
	}

	if _, err := file.WriteString("some content\n"); err != nil {
		t.Fatal(err)
	}

	file.Sync()
	file.Close()

	if mkErr1 := os.MkdirAll("./tmp/1", 0755); mkErr1 != nil {
		t.Fatal(mkErr1)
	}

	file1, fErr1 := os.Create("./tmp/1/file1_to_cache.txt")
	if fErr1 != nil {
		t.Fatal(fErr1)
	}

	if _, err := file1.WriteString("some content\n"); err != nil {
		t.Fatal(err)
	}

	file1.Sync()
	file1.Close()

	plugin := newTestPlugin("s3", true, false, []string{"./tmp/1"}, "{{ .Repo.Name }}_{{ .Commit.Branch }}_{{ .Build.Number }}", "")

	if err := plugin.Exec(); err != nil {
		t.Errorf("plugin (rebuild mode) exec failed, error: %v\n", err)
	}

	if err := os.RemoveAll("./tmp"); err != nil {
		t.Fatal(err)
	}

	plugin.Config.Rebuild = false
	plugin.Config.Restore = true
	if err := plugin.Exec(); err != nil {
		t.Errorf("plugin (restore mode) exec failed, error: %v\n", err)
	}

	if _, err := os.Stat("./tmp/1/file_to_cache.txt"); os.IsNotExist(err) {
		t.Error(err)
	}

	if _, err := os.Stat("./tmp/1/file1_to_cache.txt"); os.IsNotExist(err) {
		t.Error(err)
	}
}

func TestRestoreWithGzip(t *testing.T) {
	setup(t)
	defer cleanUp(t)

	if err := os.MkdirAll("./tmp/1", 0755); err != nil {
		t.Fatal(err)
	}

	file, cErr := os.Create("./tmp/1/file_to_cache.txt")
	if cErr != nil {
		t.Fatal(cErr)
	}

	if _, err := file.WriteString("some content\n"); err != nil {
		t.Fatal(err)
	}

	file.Sync()
	file.Close()

	if mkErr1 := os.MkdirAll("./tmp/1", 0755); mkErr1 != nil {
		t.Fatal(mkErr1)
	}

	file1, fErr1 := os.Create("./tmp/1/file1_to_cache.txt")
	if fErr1 != nil {
		t.Fatal(fErr1)
	}

	if _, err := file1.WriteString("some content\n"); err != nil {
		t.Fatal(err)
	}

	file1.Sync()
	file1.Close()

	plugin := newTestPlugin("s3", true, false, []string{"./tmp/1"}, "", "gzip")

	if err := plugin.Exec(); err != nil {
		t.Errorf("plugin (rebuild mode) exec failed, error: %v\n", err)
	}

	if err := os.RemoveAll("./tmp"); err != nil {
		t.Fatal(err)
	}

	plugin.Config.Rebuild = false
	plugin.Config.Restore = true
	if err := plugin.Exec(); err != nil {
		t.Errorf("plugin (restore mode) exec failed, error: %v\n", err)
	}

	if _, err := os.Stat("./tmp/1/file_to_cache.txt"); os.IsNotExist(err) {
		t.Error(err)
	}

	if _, err := os.Stat("./tmp/1/file1_to_cache.txt"); os.IsNotExist(err) {
		t.Error(err)
	}
}

func TestRestoreWithFilesystem(t *testing.T) {
	setup(t)
	defer cleanUp(t)

	if err := os.MkdirAll("./tmp/1", 0755); err != nil {
		t.Fatal(err)
	}

	file, cErr := os.Create("./tmp/1/file_to_cache.txt")
	if cErr != nil {
		t.Fatal(cErr)
	}

	if _, err := file.WriteString("some content\n"); err != nil {
		t.Fatal(err)
	}

	file.Sync()
	file.Close()

	if mkErr1 := os.MkdirAll("./tmp/1", 0755); mkErr1 != nil {
		t.Fatal(mkErr1)
	}

	file1, fErr1 := os.Create("./tmp/1/file1_to_cache.txt")
	if fErr1 != nil {
		t.Fatal(fErr1)
	}

	if _, err := file1.WriteString("some content\n"); err != nil {
		t.Fatal(err)
	}

	file1.Sync()
	file1.Close()

	plugin := newTestPlugin("filesystem", true, false, []string{"./tmp/1"}, "", "gzip")

	if err := plugin.Exec(); err != nil {
		t.Errorf("plugin (rebuild mode) exec failed, error: %v\n", err)
	}

	if err := os.RemoveAll("./tmp"); err != nil {
		t.Fatal(err)
	}

	plugin.Config.Rebuild = false
	plugin.Config.Restore = true
	if err := plugin.Exec(); err != nil {
		t.Errorf("plugin (restore mode) exec failed, error: %v\n", err)
	}

	if _, err := os.Stat("./tmp/1/file_to_cache.txt"); os.IsNotExist(err) {
		t.Error(err)
	}

	if _, err := os.Stat("./tmp/1/file1_to_cache.txt"); os.IsNotExist(err) {
		t.Error(err)
	}
}

// Helpers

func newTestPlugin(bck string, rebuild, restore bool, mount []string, cacheKey, archiveFmt string) Plugin {
	return Plugin{
		Logger: log.NewNopLogger(),
		Metadata: metadata.Metadata{
			Repo: metadata.Repo{
				Branch: "master",
				Name:   "drone-cache",
			},
			Commit: metadata.Commit{
				Branch: "master",
			},
		},
		Config: Config{
			ArchiveFormat:    archiveFmt,
			CompressionLevel: cache.DefaultCompressionLevel,
			Backend:          bck,
			CacheKey:         cacheKey,
			Mount:            mount,
			Rebuild:          rebuild,
			Restore:          restore,

			FileSystem: backend.FileSystemConfig{
				CacheRoot: "../testcache/cache",
			},

			S3: backend.S3Config{
				ACL:        "private",
				Bucket:     bucket,
				Encryption: "",
				Endpoint:   endpoint,
				Key:        accessKey,
				PathStyle:  true, // Should be true for minio and false for AWS.
				Region:     region,
				Secret:     secretAccessKey,
			},
		},
	}
}

func newMinioClient() (*minio.Client, error) {
	minioClient, err := minio.New(endpoint, accessKey, secretAccessKey, useSSL)
	if err != nil {
		return nil, err
	}
	return minioClient, nil
}

func setup(t *testing.T) {
	minioClient, err := newMinioClient()
	if err != nil {
		t.Fatal(err)
	}

	if err = minioClient.MakeBucket(bucket, region); err != nil {
		t.Fatal(err)
	}
}

func cleanUp(t *testing.T) {
	if err := os.RemoveAll("./tmp"); err != nil {
		t.Fatal(err)
	}

	minioClient, err := newMinioClient()
	if err != nil {
		t.Fatal(err)
	}

	if err = removeAllObjects(minioClient, bucket); err != nil {
		t.Fatal(err)
	}

	if err = minioClient.RemoveBucket(bucket); err != nil {
		t.Fatal(err)
	}
}

func removeAllObjects(minioClient *minio.Client, bucketName string) error {
	objects := make(chan string)
	errors := make(chan error)

	go func() {
		defer close(objects)
		defer close(errors)

		for object := range minioClient.ListObjects(bucketName, "", true, nil) {
			if object.Err != nil {
				errors <- object.Err
			}
			objects <- object.Key
		}
	}()

	for {
		select {
		case object, open := <-objects:
			if !open {
				return nil
			}
			if err := minioClient.RemoveObject(bucketName, object); err != nil {
				return fmt.Errorf("remove all objects failed, %v", err)
			}
		case err, open := <-errors:
			if !open {
				return nil
			}
			return fmt.Errorf("remove all objects failed, while fetching %v", err)
		}
	}
}

func getEnv(key, defaultVal string) string {
	value, ok := os.LookupEnv(key)
	if !ok {
		return defaultVal
	}
	return value
}
