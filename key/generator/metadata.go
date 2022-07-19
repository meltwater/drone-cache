package generator

import (
	"bytes"
	"crypto/md5" // #nosec
	"crypto/sha256"
	"errors"
	"fmt"
	hash2 "hash"
	"io"
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
func NewMetadata(logger log.Logger, tmpl string, data metadata.Metadata, nowFunc func() time.Time) *Metadata {
	return &Metadata{
		logger: logger,
		tmpl:   tmpl,
		data:   data,
		funcMap: template.FuncMap{
			"checksum":  checksumFunc(logger),
			"hashFiles": hashFilesFunc(logger),
			"epoch":     func() string { return strconv.FormatInt(nowFunc().Unix(), EpochNumBase) },
			"arch":      func() string { return runtime.GOARCH },
			"os":        func() string { return runtime.GOOS },
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
		return fmt.Sprintf("%x", fileHash(p, logger, md5.New))
	}
}

func hashFilesFunc(logger log.Logger) func(...string) string {
	return func(patterns ...string) string {
		var readers []io.Reader

		for _, pattern := range patterns {
			paths, err := filepath.Glob(pattern)
			if err != nil {
				level.Error(logger).Log("could not parse file path as a glob pattern")

				continue
			}

			for _, p := range paths {
				readers = append(readers, bytes.NewReader(fileHash(p, logger, sha256.New)))
			}
		}

		if len(readers) == 0 {
			level.Debug(logger).Log("no matches found for glob")

			return ""
		}

		level.Debug(logger).Log("found %d files to hash", len(readers))

		h, err := readerHasher(sha256.New, readers...)
		if err != nil {
			level.Error(logger).Log("could not generate the hash of the input files: %s", err.Error())

			return ""
		}

		return fmt.Sprintf("%x", h)
	}
}

func fileHash(path string, logger log.Logger, hasher func() hash2.Hash) []byte {
	path, err := filepath.Abs(filepath.Clean(path))
	if err != nil {
		level.Error(logger).Log("could not compute the absolute file path: %s", err.Error())

		return []byte{}
	}

	f, err := os.Open(path)
	if err != nil {
		level.Error(logger).Log("could not open the file: %s", err.Error())

		return []byte{}
	}

	defer internal.CloseWithErrLogf(logger, f, "checksum close defer")

	h, err := readerHasher(hasher, f)
	if err != nil {
		level.Error(logger).Log("could not generate the hash of the input file: %s", err.Error())

		return []byte{}
	}

	return h
}
