package autodetect

import (
	"errors"
	"fmt"
	"os"
)

type yarnPreparer struct{}

func newYarnPreparer() *yarnPreparer {
	return &yarnPreparer{}
}
func (*yarnPreparer) PrepareRepo() (string, error) {
	pathToCache := ".yarn"
	// for yarn 1.x
	err := prepareYarn(pathToCache, ".yarnrc", "\n--cache-folder %s\n")
	if err != nil {
		return "", err
	}

	// for yarn 2.x
	err = prepareYarn(pathToCache, ".yarnrc.yaml", "\ncacheFolder: \"%s\"\n")
	if err != nil {
		return "", err
	}

	return pathToCache, nil
}

func prepareYarn(pathToCache string, fileToWrite string, contentToWrite string) error {
	cmdToOverrideRepo := fmt.Sprintf(contentToWrite, pathToCache)

	if _, err := os.Stat(fileToWrite); errors.Is(err, os.ErrNotExist) {
		f, err := os.Create(fileToWrite)
		if err != nil {
			return err
		}
		defer f.Close()
		_, err = f.WriteString(cmdToOverrideRepo)

		if err != nil {
			return err
		}

		return nil
	}

	f, err := os.OpenFile(fileToWrite, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644) //nolint:gomnd

	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.WriteString(cmdToOverrideRepo)

	if err != nil {
		return err
	}

	return nil
}
