package autodetect

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/meltwater/drone-cache/test"
)

const (
	pomFile          = "pom.xml"
	nestedDirectory  = "dir"
	bazelBuildFile   = "build.gradle"
	testFileContent  = "some_content"
	testFileContent2 = "some_other_content"
	toolMaven        = "maven"
	toolMavenDir     = ".m2/repository"
	toolGradle       = "gradle"
	toolGradleDir    = ".gradle"
)

func TestDetectDirectoriesToCacheMaven(t *testing.T) {
	f, err := os.Create(pomFile)
	test.Ok(t, err)
	defer f.Close()
	_, err = f.WriteString(testFileContent)
	test.Ok(t, err)
	directoriesToCache, buildToolsDetected, hashes, err := DetectDirectoriesToCache()
	test.Ok(t, err)
	test.Ok(t, os.RemoveAll(pomFile))
	expectedCacheDir := []string{toolMavenDir}
	expectedDetectedTool := []string{toolMaven}
	test.Equals(t, directoriesToCache, expectedCacheDir)
	test.Equals(t, buildToolsDetected, expectedDetectedTool)
	test.Equals(t, hashes, "baab6c16d9143523b7865d46896e4596")
}

func TestDetectDirectoriesToCacheMavenMultiMaven(t *testing.T) {
	f, err := os.Create(pomFile)
	test.Ok(t, err)
	defer f.Close()
	_, err = f.WriteString(testFileContent)
	test.Ok(t, err)
	test.Ok(t, os.MkdirAll(nestedDirectory, 0755))
	f2, err := os.Create(filepath.Join(nestedDirectory, pomFile))
	test.Ok(t, err)
	defer f2.Close()
	_, err = f2.WriteString(testFileContent2)
	test.Ok(t, err)
	directoriesToCache, buildToolsDetected, hashes, err := DetectDirectoriesToCache()
	test.Ok(t, err)
	test.Ok(t, os.RemoveAll(pomFile))
	test.Ok(t, os.RemoveAll(filepath.Join(nestedDirectory, pomFile)))
	expectedCacheDir := []string{toolMavenDir}
	expectedDetectedTool := []string{toolMaven}
	test.Equals(t, directoriesToCache, expectedCacheDir)
	test.Equals(t, buildToolsDetected, expectedDetectedTool)
	test.Equals(t, hashes, "baab6c16d9143523b7865d46896e4596")
}

func TestDetectDirectoriesToCacheBazel(t *testing.T) {
	f, err := os.Create(bazelBuildFile)
	test.Ok(t, err)
	defer f.Close()
	_, err = f.WriteString(testFileContent)
	test.Ok(t, err)
	directoriesToCache, buildToolsDetected, hashes, err := DetectDirectoriesToCache()
	test.Ok(t, os.RemoveAll(bazelBuildFile))
	test.Ok(t, err)
	expectedCacheDir := []string{toolGradleDir}
	expectedDetectedTool := []string{toolGradle}
	test.Equals(t, directoriesToCache, expectedCacheDir)
	test.Equals(t, buildToolsDetected, expectedDetectedTool)
	test.Equals(t, hashes, "baab6c16d9143523b7865d46896e4596")
}

func TestDetectDirectoriesToCacheCombined(t *testing.T) {
	f, err := os.Create(bazelBuildFile)
	test.Ok(t, err)
	defer f.Close()
	_, err = f.WriteString(testFileContent)
	test.Ok(t, err)
	f2, err := os.Create(pomFile)
	test.Ok(t, err)
	defer f2.Close()
	_, err = f2.WriteString(testFileContent2)
	test.Ok(t, err)
	directoriesToCache, buildToolsDetected, hashes, err := DetectDirectoriesToCache()
	test.Ok(t, os.RemoveAll(bazelBuildFile))
	test.Ok(t, os.RemoveAll(pomFile))
	test.Ok(t, err)
	expectedCacheDir := []string{toolMavenDir, toolGradleDir}
	expectedDetectedTool := []string{toolMaven, toolGradle}
	test.Equals(t, directoriesToCache, expectedCacheDir)
	test.Equals(t, buildToolsDetected, expectedDetectedTool)
	test.Equals(t, hashes, "1eb00e74bffac0c4fa2d6dbfd8c26cb7baab6c16d9143523b7865d46896e4596")
}
