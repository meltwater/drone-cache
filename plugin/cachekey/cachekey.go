package cachekey

import (
	"crypto/md5" // #nosec
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/meltwater/drone-cache/metadata"
)

var funcMap = template.FuncMap{
	"checksum": func(path string) string {
		absPath, err := filepath.Abs(filepath.Clean(path))
		if err != nil {
			log.Println("cache key template/checksum could not find file")
			return ""
		}

		f, err := os.Open(absPath)
		if err != nil {
			log.Println("cache key template/checksum could not open file")
			return ""
		}
		defer f.Close()

		str, err := readerHasher(f)
		if err != nil {
			log.Println("cache key template/checksum could not generate hash")
			return ""
		}
		return str
	},
	"epoch": func() string { return strconv.FormatInt(time.Now().Unix(), 10) },
	"arch":  func() string { return runtime.GOARCH },
	"os":    func() string { return runtime.GOOS },
}

// Generate generates key from given template as parameter or fallbacks hash
func Generate(tmpl, mount string, data metadata.Metadata) (string, error) {
	if tmpl == "" {
		return "", errors.New("cache key template is empty")
	}

	t, err := ParseTemplate(tmpl)
	if err != nil {
		return "", fmt.Errorf("could not parse <%s> as cache key template, falling back to default %w", tmpl, err)
	}

	var b strings.Builder

	err = t.Execute(&b, data)
	if err != nil {
		return "", fmt.Errorf("could not build <%s> as cache key, falling back to default %w", tmpl, err)
	}

	return filepath.Join(b.String(), mount), nil
}

// ParseTemplate parses and mounts helper functions to template engine
func ParseTemplate(tmpl string) (*template.Template, error) {
	return template.New("cacheKey").Funcs(funcMap).Parse(tmpl)
}

// Hash generates a key based on given strings (ie. filename paths and branch)
func Hash(parts ...string) (string, error) {
	readers := make([]io.Reader, len(parts))
	for i, p := range parts {
		readers[i] = strings.NewReader(p)
	}

	return readerHasher(readers...)
}

// Helpers

// readerHasher generic md5 hash generater from io.Readers
func readerHasher(readers ...io.Reader) (string, error) {
	h := md5.New() // #nosec

	for _, r := range readers {
		if _, err := io.Copy(h, r); err != nil {
			return "", fmt.Errorf("could not write reader as hash %w", err)
		}
	}

	return fmt.Sprintf("%x", h.Sum(nil)), nil
}
