package plugin

import (
	"fmt"
	"os"
	"testing"

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

	if mkErr1 := os.MkdirAll("./tmp/1", 0755); mkErr1 != nil {
		t.Fatal(mkErr1)
	}

	file, fErr := os.Create("./tmp/1/file_to_cache.txt")
	if fErr != nil {
		t.Fatal(fErr)
	}

	_, wErr := file.WriteString("some content\n")
	if wErr != nil {
		t.Fatal(wErr)
	}
	file.Sync()
	file.Close()

	plugin := newTestPlugin(true, false, []string{"./tmp/1"})

	if err := plugin.Exec(); err != nil {
		t.Errorf("plugin exec failed, error: %v\n", err)
	}
}

func TestRestore(t *testing.T) {
	setup(t)
	defer cleanUp(t)

	if err := os.MkdirAll("./tmp/1", 0755); err != nil {
		t.Fatal(err)
	}

	file, cErr := os.Create("./tmp/1/file_to_cache.txt")
	if cErr != nil {
		t.Fatal(cErr)
	}

	_, wErr := file.WriteString("some content\n")
	if wErr != nil {
		t.Fatal(wErr)
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

	_, wErr1 := file1.WriteString("some content\n")
	if wErr1 != nil {
		t.Fatal(wErr1)
	}

	file1.Sync()
	file1.Close()

	plugin := newTestPlugin(true, false, []string{"./tmp/1"})

	if err := plugin.Exec(); err != nil {
		t.Errorf("plugin (rebuild mode) exec failed, error: %v\n", err)
	}

	if err := os.RemoveAll("./tmp"); err != nil {
		t.Fatal(err)
	}

	plugin.Rebuild = false
	plugin.Restore = true
	if err := plugin.Exec(); err != nil {
		t.Errorf("plugin (restore mode) exec failed, error: %v\n", err)
	}

	if _, err := os.Stat("./tmp/1/file_to_cache.txt"); os.IsNotExist(err) {
		t.Fatal(err)
	}
}

// Helpers

func newTestPlugin(rebuild bool, restore bool, mount []string) Plugin {
	return Plugin{
		Config: Config{
			ACL:        "private",
			Branch:     "master",
			Bucket:     bucket,
			Default:    "master",
			Encryption: "",
			Endpoint:   endpoint,
			Key:        accessKey,
			Mount:      mount,
			PathStyle:  true, // Should be true for minio and false for AWS.
			Rebuild:    rebuild,
			Region:     region,
			Repo:       "drone-s3-cache",
			Restore:    restore,
			Secret:     secretAccessKey,
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
			if !open { // Unlikely to happend, I guess, still learning!
				return nil
			}
			return fmt.Errorf("remove all objects failed, while fetching %v", err)
		}
	}

	return nil
}

func getEnv(key, defaultVal string) string {
	value, ok := os.LookupEnv(key)
	if !ok {
		return defaultVal
	}
	return value
}
