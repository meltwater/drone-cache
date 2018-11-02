package plugin

import (
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/minio/minio-go"
)

const (
	endpoint        = "127.0.0.1:9000"
	accessKeyId     = "AKIAIOSFODNN7EXAMPLE"
	secretAccessKey = "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY"
	bucket          = "meltwater-drone-test"
	region          = "eu-west-1"
	useSSL          = false
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
		Rebuild:    rebuild,
		Restore:    restore,
		Mount:      mount,
		Endpoint:   endpoint,
		Key:        accessKeyId,
		Secret:     secretAccessKey,
		Bucket:     bucket,
		Region:     region,
		ACL:        "private",
		Encryption: "",
		PathStyle:  true, // Should be true for minio and false for AWS.
		Repo:       "drone-s3-cache",
		Default:    "master",
		Branch:     "master",
	}
}

func newMinioClient() (*minio.Client, error) {
	minioClient, err := minio.New(endpoint, accessKeyId, secretAccessKey, useSSL)
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
