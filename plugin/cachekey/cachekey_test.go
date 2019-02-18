package cachekey

import (
	"strings"
	"testing"

	"github.com/meltwater/drone-cache/metadata"
)

func TestGenerate(t *testing.T) {
	actual, err := Generate("{{ .Repo.Name }}", "", metadata.Metadata{})
	if err != nil {
		t.Errorf("generate failed, error: %v\n", err)
	}

	expected := ""
	if actual != expected {
		t.Errorf("generate failed, got: %s, want: %s\n", actual, expected)
	}
}

func TestParseTemplate(t *testing.T) {
	tmpl, err := ParseTemplate("tmpl")
	if err != nil {
		t.Errorf("parser template failed, error: %v\n", err)
	}

	var b strings.Builder
	err = tmpl.Execute(&b, metadata.Metadata{})
	if err != nil {
		t.Errorf("parser template failed, error: %v\n", err)
	}

	actual := b.String()
	expected := "tmpl"
	if actual != expected {
		t.Errorf("parse template failed, got: %s, want: %s\n", actual, expected)
	}
}

func TestHash(t *testing.T) {
	actual, err := Hash("")
	if err != nil {
		t.Errorf("hash failed, error: %v\n", err)
	}

	expected := "d41d8cd98f00b204e9800998ecf8427e"
	if actual != expected {
		t.Errorf("hash failed, got: %s, want: %s\n", actual, expected)
	}
}
