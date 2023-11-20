package plugin

import (
	"fmt"
	"github.com/meltwater/drone-cache/test"
	"log"
	"os"
	"testing"
)

func TestExpandConfigPath_Tilde(t *testing.T) {
	testMountTilde := []string{"~/test/path"}
	wantExpandedMountTilde := []string{fmt.Sprintf("%s/test/path", osExpand("$HOME"))}
	gotExpandedMountTilde := expandConfigPath(testMountTilde)
	test.Equals(t, wantExpandedMountTilde, gotExpandedMountTilde)

}

func TestExpandConfigPath_EnvVar(t *testing.T) {
	testMount := []string{"$HOME/test/path"}
	wantExpandedMount := []string{fmt.Sprintf("%s/test/path", osExpand("$HOME"))}
	gotExpandedMount := expandConfigPath(testMount)
	test.Equals(t, wantExpandedMount, gotExpandedMount)
}

func osExpand(symbol string) string {
	absolutePath := os.ExpandEnv(symbol)
	if absolutePath == "" {
		log.Fatalf("Could not find the absolute path for symbol/envVar: %s", symbol)
	}
	return absolutePath
}
