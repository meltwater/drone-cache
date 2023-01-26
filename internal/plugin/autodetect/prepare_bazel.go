package autodetect

import (
	"errors"
	"fmt"
	"os"
)

type bazelPreparer struct{}

func newBazelPreparer() *bazelPreparer {
	return &bazelPreparer{}
}

func (*bazelPreparer) PrepareRepo() (string, error) {
	fileName := ".bazelrc"
	pathToCache := ".bazel"
	cmdToOverrideRepo := fmt.Sprintf("build --test_tmpdir=%s\ntest --test_tmpdir=%s", pathToCache, pathToCache)

	if _, err := os.Stat(fileName); errors.Is(err, os.ErrNotExist) {
		f, err := os.Create(fileName)
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

	f, err := os.OpenFile(fileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)

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
