package generator

import (
	"testing"
	"text/template"

	"github.com/go-kit/kit/log"
	"github.com/meltwater/drone-cache/internal/metadata"
	"github.com/meltwater/drone-cache/test"
)

func TestGenerate(t *testing.T) {
	t.Parallel()

	l := log.NewNopLogger()

	for _, tt := range []struct {
		given    string
		expected string
	}{
		{`{{ .Repo.Name }}`, "RepoName"},
		{`{{ checksum "checksum_file_test.txt"}}`, "04a29c732ecbce101c1be44c948a50c6"},
		{`{{ checksum "../../docs/drone_env_vars.md"}}`, "f8b5b7f96f3ffaa828e4890aab290e59"},
		{`{{ epoch }}`, "1550563151"},
		{`{{ arch }}`, "amd64"},
		{`{{ os }}`, "darwin"},
	} {
		tt := tt
		t.Run(tt.given, func(t *testing.T) {
			g := Metadata{
				logger: l,
				tmpl:   tt.given,
				data:   metadata.Metadata{Repo: metadata.Repo{Name: "RepoName"}},
				funcMap: template.FuncMap{
					"checksum": checksumFunc(l),
					"epoch":    func() string { return "1550563151" },
					"arch":     func() string { return "amd64" },
					"os":       func() string { return "darwin" },
				},
			}

			actual, err := g.Generate(tt.given)
			test.Ok(t, err)
			test.Equals(t, actual, tt.expected)
		})
	}
}

func TestParseTemplate(t *testing.T) {
	t.Parallel()

	l := log.NewNopLogger()

	for _, tt := range []struct {
		given string
	}{
		{`{{ .Repo.Name }}`},
		{`{{ checksum "checksum_file_test.txt"}}`},
		{`{{ epoch }}`},
		{`{{ arch }}`},
		{`{{ os }}`},
	} {
		tt := tt
		t.Run(tt.given, func(t *testing.T) {
			g := Metadata{
				logger: l,
				tmpl:   tt.given,
				data:   metadata.Metadata{Repo: metadata.Repo{Name: "RepoName"}},
				funcMap: template.FuncMap{
					"checksum": checksumFunc(l),
					"epoch":    func() string { return "1550563151" },
					"arch":     func() string { return "amd64" },
					"os":       func() string { return "darwin" },
				},
			}

			_, err := g.parseTemplate()
			test.Ok(t, err)
		})
	}
}
