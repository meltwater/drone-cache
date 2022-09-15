// +build integration

package sftp

import (
	"bytes"
	"context"
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/meltwater/drone-cache/test"

	"github.com/go-kit/log"
)

const (
	defaultSFTPHost  = "127.0.0.1"
	defaultSFTPPort  = "22"
	defaultUsername  = "foo"
	defaultPassword  = "pass"
	defaultCacheRoot = "/sftp_test"
)

var (
	host      = getEnv("TEST_SFTP_HOST", defaultSFTPHost)
	port      = getEnv("TEST_SFTP_PORT", defaultSFTPPort)
	username  = getEnv("TEST_SFTP_USERNAME", defaultUsername)
	password  = getEnv("TEST_SFTP_PASSWORD", defaultPassword)
	cacheRoot = getEnv("TEST_SFTP_CACHE_ROOT", defaultCacheRoot)
)

func TestRoundTrip(t *testing.T) {
	t.Parallel()

	backend, cleanUp := setup(t)
	t.Cleanup(cleanUp)

	content := "Hello world4"

	// Test Put
	test.Ok(t, backend.Put(context.TODO(), "test.t", strings.NewReader(content)))

	// Test Get
	var buf bytes.Buffer
	test.Ok(t, backend.Get(context.TODO(), "test.t", &buf))

	b, err := ioutil.ReadAll(&buf)
	test.Ok(t, err)

	test.Equals(t, []byte(content), b)

	exists, err := backend.Exists(context.TODO(), "test.t")
	test.Ok(t, err)

	test.Equals(t, true, exists)
}

// Helpers

func setup(t *testing.T) (*Backend, func()) {
	b, err := New(
		log.NewNopLogger(),
		Config{
			CacheRoot: cacheRoot,
			Username:  username,
			Auth: SSHAuth{
				Password: password,
				Method:   SSHAuthMethodPassword,
			},
			Host: host,
			Port: port,
		},
	)
	test.Ok(t, err)

	return b, func() {}
}

func getEnv(key, defaultVal string) string {
	value, ok := os.LookupEnv(key)
	if !ok {
		return defaultVal
	}

	return value
}
