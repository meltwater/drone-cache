package plugin

import (
	"fmt"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/meltwater/drone-cache/archive"
	"github.com/meltwater/drone-cache/test"
)

const (
	testRoot                       = "testdata"
	defaultStorageOperationTimeout = 5 * time.Second
)

func TestHandleMount(t *testing.T) {
	test.Ok(t, os.Mkdir(testRoot, 0755))
	t.Cleanup(func() {
		os.RemoveAll(testRoot)
	})
	cases := []struct {
		name           string
		mounts         []string
		expectedMounts []string
		makeFiles      func()
	}{
		{
			name:           "handle-mount-single",
			mounts:         []string{"test/single"},
			expectedMounts: []string{"test/single"},
			makeFiles:      func() {},
		},
		{
			name:           "handle-mount-nested",
			mounts:         []string{"test/a", "test/b"},
			expectedMounts: []string{"test/a", "test/b"},
			makeFiles:      func() {},
		},
		{
			name:           "handle-mount-glob-empty",
			mounts:         []string{"test/**", "test/b"},
			expectedMounts: []string{"test/b"},
			makeFiles:      func() {},
		},
		{
			name:   "handle-mount-glob-notempty",
			mounts: []string{fmt.Sprintf("%s/%s", testRoot, "test/**")},
			expectedMounts: []string{
				fmt.Sprintf("%s/%s", testRoot, "test/nestedA"),
				fmt.Sprintf("%s/%s", testRoot, "test/nestedB"),
			},
			makeFiles: func() {
				// Make test directories for glob to work properly
				os.MkdirAll(fmt.Sprintf("%s/%s", testRoot, "test/nestedA"), 0755)
				os.MkdirAll(fmt.Sprintf("%s/%s", testRoot, "test/nestedB"), 0755)
			},
		},
	}

	for _, tc := range cases {
		c := defaultConfig()
		c.Mount = tc.mounts

		tc.makeFiles()
		test.Ok(t, c.HandleMount())

		test.Assert(t, reflect.DeepEqual(c.Mount, tc.expectedMounts),
			"expected mount differs from handled mount result:\nexpected: %v\ngot:%v", tc.expectedMounts, c.Mount)
	}
}

// Config plugin configuration

func defaultConfig() *Config {
	return &Config{
		CompressionLevel:        archive.DefaultCompressionLevel,
		StorageOperationTimeout: defaultStorageOperationTimeout,
		Override:                true,
	}
}

func containsGlob(a []string) bool {
	for _, v := range a {
		if strings.Contains(v, "**") {
			return true
		}
	}

	return false
}
