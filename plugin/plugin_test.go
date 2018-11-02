package plugin

import (
	"fmt"
	"log"
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

func TestRebuild(t *testing.T) {
	setup(t)
	defer cleanUp(t)

	if mkErr1 := os.MkdirAll("./tmp/1", 0755); mkErr1 != nil {
		t.Fatal(mkErr1)
	}

	file, ferr := os.Create("./tmp/1/file_to_cache.txt")
	if ferr != nil {
		t.Fatal(ferr)
	}

	_, werr := file.WriteString("some content\n")
	if werr != nil {
		t.Fatal(werr)
	}
	file.Sync()
	file.Close()

	plugin := newTestPlugin(true, false, []string{"./tmp/1"})

	if err := plugin.Exec(); err != nil {
		t.Errorf("plugin exec failed, error: %v\n", err)
	}

	// TODO: Move as clean up
	if rErr := os.RemoveAll("./tmp"); rErr != nil {
		t.Fatal(rErr)
	}
}

func TestRestore(t *testing.T) {
	setup(t)
	defer cleanUp(t)

	if mkErr := os.MkdirAll("./tmp/1", 0755); mkErr != nil {
		t.Fatal(mkErr)
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

	file1, cErr1 := os.Create("./tmp/1/file1_to_cache.txt")
	if cErr1 != nil {
		t.Fatal(cErr1)
	}

	_, wErr1 := file1.WriteString("some content\n")
	if wErr1 != nil {
		t.Fatal(wErr1)
	}

	file1.Sync()
	file1.Close()

	plugin := newTestPlugin(true, false, []string{"./tmp/1"})

	if xErr := plugin.Exec(); xErr != nil {
		t.Errorf("plugin (rebuild mode) exec failed, error: %v\n", xErr)
	}

	if rErr := os.RemoveAll("./tmp"); rErr != nil {
		t.Fatal(rErr)
	}

	plugin.Rebuild = false
	plugin.Restore = true
	if err := plugin.Exec(); err != nil {
		t.Errorf("plugin (restore mode) exec failed, error: %v\n", err)
	}

	if _, err := os.Stat("./tmp/1/file_to_cache.txt"); os.IsNotExist(err) {
		t.Fatal(err)
	}

	// TODO: Move as clean up
	if rErr := os.RemoveAll("./tmp"); rErr != nil {
		t.Fatal(rErr)
	}
}

// Helpers

func newTestPlugin(rebuild bool, restore bool, mount []string) Plugin {
	return Plugin{
		ACL:        "private",
		Branch:     "master",
		Bucket:     bucket,
		Default:    "master",
		Encryption: "",
		Endpoint:   endpoint(),
		Key:        accessKey(),
		Mount:      mount,
		PathStyle:  true, // Should be true for minio and false for AWS.
		Rebuild:    rebuild,
		Region:     region,
		Repo:       "drone-s3-cache",
		Restore:    restore,
		Secret:     secretAccessKey(),
	}
}

func newMinioClient() (*minio.Client, error) {
	minioClient, err := minio.New(endpoint(), accessKey(), secretAccessKey(), useSSL)
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
	objectsCh := make(chan string)

	go func() {
		defer close(objectsCh)

		for object := range minioClient.ListObjects(bucketName, "", true, nil) {
			if object.Err != nil {
				// TODO: Log statement!
				log.Fatalln(object.Err)
			}
			objectsCh <- object.Key
		}
	}()

	for rErr := range minioClient.RemoveObjects(bucketName, objectsCh) {
		return fmt.Errorf("remove all objects failed, %v", rErr)
	}

	return nil
}

func endpoint() string {
	value, ok := os.LookupEnv("TEST_ENDPOINT")
	if !ok {
		return defaultEndpoint
	}
	return value
}

func accessKey() string {
	value, ok := os.LookupEnv("TEST_ACCESS_KEY")
	if !ok {
		return defaultAccessKey
	}
	return value
}

func secretAccessKey() string {
	value, ok := os.LookupEnv("TEST_SECRET_KEY")
	if !ok {
		return defaultSecretAccessKey
	}
	return value
}
