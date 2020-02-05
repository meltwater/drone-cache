package backend

import (
	"bytes"
	"io/ioutil"
	"math/rand"
	"testing"

	"github.com/go-kit/kit/log"
)

const defaultBlobStorageURL = "127.0.0.1:10000"

var blobURL = getEnv("TEST_AZURITE_URL", defaultBlobStorageURL)

func TestAzureTruth(t *testing.T) {

	b, err := InitializeAzureBackend(log.NewNopLogger(),
		AzureConfig{
			AccountName:    "devstoreaccount1",
			AccountKey:     "Eby8vdM02xNOcqFlqUwJPLlmEtlCDXJ1OUzFT50uSRZ6IFsuFq2UVErCz4I6tq/K1SZFPTOtr/KBHBeksoGMGw==",
			ContainerName:  "testcontainer",
			BlobStorageURL: blobURL,
			Azurite:        true,
		}, true)
	if err != nil {
		t.Fatal(err)
	}

	token := make([]byte, 32)
	rand.Read(token)
	testData := bytes.NewReader(token)

	// PUT TEST
	err = b.Put("test_key", testData)
	if err != nil {
		t.Fatal(err)
	}

	// GET TEST
	readCloser, err := b.Get("test_key")
	if err != nil {
		t.Fatal(err)
	}

	// Check the validity of returned bytes
	readData, _ := ioutil.ReadAll(readCloser)

	if !bytes.Equal(readData, token) {
		t.Fatal(string(readData), "!=", token)
	}

}
