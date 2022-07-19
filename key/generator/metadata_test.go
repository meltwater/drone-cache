package generator

import (
	"runtime"
	"testing"
	"time"

	"github.com/go-kit/log"
	"github.com/meltwater/drone-cache/internal/metadata"
	"github.com/meltwater/drone-cache/test"
)

func TestGenerate(t *testing.T) {
	t.Parallel()

	logger := log.NewNopLogger()

	for _, tt := range []struct {
		given    string
		expected string
	}{
		{`{{ .Repo.Name }}`, "RepoName"},
		{`{{ checksum "checksum_file_test.txt"}}`, "04a29c732ecbce101c1be44c948a50c6"},
		{`{{ checksum "../../docs/drone_env_vars.md"}}`, "f8b5b7f96f3ffaa828e4890aab290e59"},
		{`{{ hashFiles "" }}`, ""},
		{`{{ hashFiles "checksum_file_test.txt" }}`, "b9fff559e00dd879bdffee979ee73e08c67dee2117da071083d3b833cbff7bc8"},
		{`{{ hashFiles "checksum_file_test.txt" "checksum_file_test.txt" }}`, "fed16eb2e98f501968c74e261feb26a8776b2ae03b205ad7302f949e75ca455f"},
		{`{{ hashFiles "checksum_file_tes*.txt" }}`, "b9fff559e00dd879bdffee979ee73e08c67dee2117da071083d3b833cbff7bc8"},
		{`{{ epoch }}`, "1550563151"},
		{`{{ arch }}`, runtime.GOARCH},
		{`{{ os }}`, runtime.GOOS},
	} {
		tt := tt
		t.Run(tt.given, func(t *testing.T) {
			g := NewMetadata(
				logger,
				tt.given,
				metadata.Metadata{Repo: metadata.Repo{Name: "RepoName"}},
				func() time.Time {
					return time.Unix(1550563151, 0)
				},
			)

			actual, err := g.Generate(tt.given)
			test.Ok(t, err)
			test.Equals(t, tt.expected, actual)
		})
	}
}
