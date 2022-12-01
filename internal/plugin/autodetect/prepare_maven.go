package autodetect

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

type mavenPreparer struct{}

func newMavenPreparer() *mavenPreparer {
	return &mavenPreparer{}
}
func (*mavenPreparer) PrepareRepo() (string, error) {
	configPath := filepath.Join(".mvn")
	fileName := "maven.config"
	pathToCache := filepath.Join(".m2", "repository")
	cmdToOverrideRepo := fmt.Sprintf(" -Dmaven.repo.local=%s ", pathToCache)

	if _, err := os.Stat(filepath.Join(configPath, fileName)); errors.Is(err, os.ErrNotExist) {
		err := os.MkdirAll(configPath, os.ModePerm)

		if err != nil {
			return "", err
		}

		f, err := os.Create(filepath.Join(configPath, fileName))

		if err != nil {
			return "", err
		}
		defer f.Close()
		_, err = f.WriteString(cmdToOverrideRepo)

		if err != nil {
			return "", err
		}

		return pathToCache, err
	}

	f, err := os.OpenFile(configPath+fileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)

	if err != nil {
		return "", err
	}
	defer f.Close()
	_, err = f.WriteString(cmdToOverrideRepo)

	if err != nil {
		return "", err
	}

	return pathToCache, nil
}
