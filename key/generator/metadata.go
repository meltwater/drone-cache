package generator

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/meltwater/drone-cache/internal"
	"github.com/meltwater/drone-cache/internal/metadata"
)

const (
	EpochNumBase = 10
)

// Metadata implements a key generator that uses a template engine to generate keys from give metadata.
type Metadata struct {
	logger log.Logger

	tmpl    string
	data    metadata.Metadata
	funcMap template.FuncMap
}

// NewMetadata creates a new Key Generator.
func NewMetadata(logger log.Logger, tmpl string, data metadata.Metadata) *Metadata {
	return &Metadata{
		logger: logger,
		tmpl:   tmpl,
		data:   data,
		funcMap: template.FuncMap{
			"checksum": checksumFunc(logger),
			"epoch":    func() string { return strconv.FormatInt(time.Now().Unix(), EpochNumBase) },
			"arch":     func() string { return runtime.GOARCH },
			"os":       func() string { return runtime.GOOS },
		},
	}
}

// Generate generates key from given template as parameter or fallbacks hash.
func (g *Metadata) Generate(_ ...string) (string, error) {
	// NOTICE: for now only consume a single template which will be changed.
	level.Info(g.logger).Log("msg", "using provided cache key template")

	if g.tmpl == "" {
		return "", errors.New("cache key template is empty")
	}

	t, err := g.parseTemplate()
	if err != nil {
		return "", fmt.Errorf("parse, <%s> as cache key template, falling back to default, %w", g.tmpl, err)
	}

	var b strings.Builder

	err = t.Execute(&b, g.data)
	if err != nil {
		return "", fmt.Errorf("build, <%s> as cache key, falling back to default, %w", g.tmpl, err)
	}

	return b.String(), nil
}

// Check checks if template is parsable.
func (g *Metadata) Check() error {
	_, err := g.parseTemplate()

	return err
}

// Helpers

func (g *Metadata) parseTemplate() (*template.Template, error) {
	tmpl, err := template.New("cacheKey").Funcs(g.funcMap).Parse(g.tmpl)
	if err != nil {
		return &template.Template{}, fmt.Errorf("parse template failed, %w", err)
	}

	return tmpl, nil
}

func checksumFunc(logger log.Logger) func(string) string {
	return func(p string) string {
		path, err := filepath.Abs(filepath.Clean(p))
		if err != nil {
			level.Error(logger).Log("cache key template/checksum could not find file")

			return ""
		}

		f, err := os.Open(path)
		if err != nil {
			level.Error(logger).Log("cache key template/checksum could not open file")

			return ""
		}

		defer internal.CloseWithErrLogf(logger, f, "checksum close defer")

		str, err := readerHasher(f)
		if err != nil {
			level.Error(logger).Log("cache key template/checksum could not generate hash")

			return ""
		}

		return str
	}
}
