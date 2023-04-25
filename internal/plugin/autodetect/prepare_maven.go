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
func (*mavenPreparer) PrepareRepo(dir string) (string, error) {
	configPath := filepath.Join(dir, ".mvn")
	fileName := "maven.config"
	pathToCache := filepath.Join(dir, ".m2", "repository")
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

	f, err := os.OpenFile(filepath.Join(configPath, fileName), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644) //nolint:gomnd

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
