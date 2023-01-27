package autodetect

import (
	"errors"
	"fmt"
	"os"
)

type gradlePreparer struct{}

func newGradlePreparer() *gradlePreparer {
	return &gradlePreparer{}
}

func (*gradlePreparer) PrepareRepo() (string, error) {
	fileName := "gradle.properties"
	pathToCache := ".gradle"
	cmdToOverrideRepo := fmt.Sprintf("systemProp.gradle.user.home=/%s/\norg.gradle.caching=true\n", pathToCache)

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

	f, err := os.OpenFile(fileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644) //nolint:gomnd

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
