package auto_detect

import (
	"crypto/md5"
	"encoding/hex"
	"io"
	"os"
	"path/filepath"
)

type buildToolInfo struct {
	globToDetect string
	tool         string
	dirToCache   string
	preparer     RepoPreparer
}

var buildToolInfoMapping = []buildToolInfo{
	{
		globToDetect: "*pom.xml",
		tool:         "maven",
		preparer:     newMavenPreparer(),
	},
	{
		globToDetect: "*build.gradle",
		tool:         "gradle",
		preparer:     newGradlePreparer(),
	},
}

func AutoDetectDirectoriesToCache() ([]string, []string, string, error) {
	var directoriesToCache []string
	var buildToolsDetected []string
	var hashes string
	for _, supportedTool := range buildToolInfoMapping {
		hash, err := hashIfFileExist(supportedTool.globToDetect)
		if err != nil {
			return nil, nil, "", err
		}
		if hash != "" {
			dirToCache, err := supportedTool.preparer.PrepareRepo()
			if err != nil {
				return nil, nil, "", err
			}
			directoriesToCache = append(directoriesToCache, dirToCache)
			buildToolsDetected = append(buildToolsDetected, supportedTool.tool)
			hashes += hash
		}
	}

	return directoriesToCache, buildToolsDetected, hashes, nil
}

func hashIfFileExist(glob string) (string, error) {
	matches, _ := filepath.Glob(glob)
	var found []string
	for _, match := range matches {
		found = append(found, match)
	}
	if len(found) == 0 {
		return "", nil
	}
	return calculateMd5FromFiles(found)
}

func calculateMd5FromFiles(fileList []string) (string, error) {
	rootMostFile := shortestPath(fileList)
	file, err := os.Open(rootMostFile)

	if err != nil {
		return "", err
	}

	defer file.Close()
	if err != nil {
		return "", err
	}

	hash := md5.New()
	_, err = io.Copy(hash, file)

	if err != nil {
		return "", err
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}

func shortestPath(input []string) (shortest string) {
	size := len(input[0])
	for _, v := range input {
		if len(v) <= size {
			shortest = v
			size = len(v)
		}
	}
	return
}
