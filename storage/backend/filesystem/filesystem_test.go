package filesystem

import (
	"bytes"
	"context"
	"io/ioutil"
	"strings"
	"testing"

	"github.com/go-kit/kit/log"
	"github.com/meltwater/drone-cache/test"
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
}

// Helpers

func setup(t *testing.T) (*Backend, func()) {
	dir, cleanUp := test.CreateTempDir(t, "filesystem-test")

	b, err := New(
		log.NewNopLogger(),
		Config{CacheRoot: dir},
	)
	test.Ok(t, err)

	return b, func() { cleanUp() }
}
